package openrouter

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	transcriptionAPI = "https://openrouter.ai/api/v1/audio/transcriptions"
	speechAPI        = "https://openrouter.ai/api/v1/audio/speech"
	defaultAudioMime = "audio/mpeg"
)

type speechModels struct {
	Data []struct {
		ID              string   `json:"id"`
		SupportedVoices []string `json:"supported_voices"`
	} `json:"data"`
}

func (a *Agent) voices(ctx context.Context) ([]string, error) {
	data, code, err := go_pkg_http.GET[speechModels](ctx, a.httpClient, modelsAPI+"?output_modalities=speech", a.headers())
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: %w", err)
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("openrouter speech models: http %d", code)
	}
	for _, m := range data.Data {
		if m.ID == a.model {
			return m.SupportedVoices, nil
		}
	}
	return nil, nil
}

func (a *Agent) Transcribe(ctx context.Context, audio []byte, opts llmrouter.STTOptions) (*llmrouter.STTResult, error) {
	if len(audio) == 0 {
		return nil, fmt.Errorf("openrouter.Transcribe: audio is empty")
	}

	mime := opts.MimeType
	if mime == "" {
		mime = defaultAudioMime
	}
	body := map[string]any{
		"model": a.model,
		"file":  go_pkg_http.File{Name: "audio", ContentType: mime, Data: audio},
	}
	if opts.Language != "" {
		body["language"] = opts.Language
	}
	if opts.Prompt != "" {
		body["prompt"] = opts.Prompt
	}

	headers := a.headers()
	delete(headers, "Content-Type")
	result, code, err := go_pkg_http.POST[struct {
		Text string `json:"text"`
	}](ctx, a.httpClient, transcriptionAPI, headers, body, "multipart")
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: POST: %w", err)
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("openrouter transcriptions: http %d", code)
	}
	return &llmrouter.STTResult{Text: strings.TrimSpace(result.Text)}, nil
}

func (a *Agent) Speak(ctx context.Context, text string, opts llmrouter.TTSOptions) (*llmrouter.TTSResult, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("openrouter.Speak: text is empty")
	}

	format := "pcm"
	if strings.EqualFold(strings.TrimSpace(opts.Format), "mp3") {
		format = "mp3"
	}
	body := map[string]any{
		"model":           a.model,
		"input":           text,
		"response_format": format,
	}

	voices, err := a.voices(ctx)
	if err != nil {
		return nil, err
	}
	voice := opts.Voice
	switch {
	case voice == "" && len(voices) > 0:
		voice = voices[0]
	case voice != "" && len(voices) > 0 && !slices.Contains(voices, voice):
		return nil, fmt.Errorf("openrouter.Speak: voice %q is not supported by %s; supported voices: %s", voice, a.model, strings.Join(voices, ", "))
	}
	if voice != "" {
		body["voice"] = voice
	}

	resp, err := go_pkg_http.POSTStream(ctx, a.httpClient, speechAPI, a.headers(), body, "json")
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: POSTStream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, llmrouter.ErrorBodyLimit))
		return nil, fmt.Errorf("openrouter speech: http %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	audio, err := io.ReadAll(io.LimitReader(resp.Body, llmrouter.StreamBodyLimit))
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}
	mime := resp.Header.Get("Content-Type")
	if format == "pcm" {
		return &llmrouter.TTSResult{Audio: llmrouter.WrapPCM16(audio, llmrouter.PCMRate(mime)), MimeType: "audio/wav"}, nil
	}
	if mime == "" {
		mime = defaultAudioMime
	}
	return &llmrouter.TTSResult{Audio: audio, MimeType: mime}, nil
}
