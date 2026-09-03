package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
)

const (
	transcriptionAPI = "https://api.openai.com/v1/audio/transcriptions"
	speechAPI        = "https://api.openai.com/v1/audio/speech"
	defaultVoice     = "alloy"
	speechFormat     = "wav"
	speechMime       = "audio/wav"
)

func (a *Agent) Transcribe(ctx context.Context, audio []byte, opts core.STTOptions) (*core.STTResult, error) {
	if len(audio) == 0 {
		return nil, fmt.Errorf("openai.Transcribe: audio is empty")
	}

	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	part, err := form.CreateFormFile("file", "audio"+audioExt(opts.MimeType))
	if err != nil {
		return nil, fmt.Errorf("multipart.CreateFormFile: %w", err)
	}
	if _, err := part.Write(audio); err != nil {
		return nil, fmt.Errorf("part.Write: %w", err)
	}
	fields := map[string]string{"model": a.model}
	if opts.Language != "" {
		fields["language"] = opts.Language
	}
	if opts.Prompt != "" {
		fields["prompt"] = opts.Prompt
	}
	for key, value := range fields {
		if err := form.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("form.WriteField: %w", err)
		}
	}
	if err := form.Close(); err != nil {
		return nil, fmt.Errorf("form.Close: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, transcriptionAPI, &body)
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", form.FormDataContentType())

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("httpClient.Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return nil, fmt.Errorf("openai transcriptions: http %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var out struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("json.Decode: %w", err)
	}
	return &core.STTResult{Text: strings.TrimSpace(out.Text)}, nil
}

func (a *Agent) Speak(ctx context.Context, text string, opts core.TTSOptions) (*core.TTSResult, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("openai.Speak: text is empty")
	}

	voice := opts.Voice
	if voice == "" {
		voice = defaultVoice
	}
	payload, err := json.Marshal(map[string]any{
		"model":           a.model,
		"input":           text,
		"voice":           voice,
		"response_format": speechFormat,
	})
	if err != nil {
		return nil, fmt.Errorf("json.Marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, speechAPI, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("httpClient.Do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return nil, fmt.Errorf("openai speech: http %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	audio, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("io.ReadAll: %w", err)
	}
	mime := resp.Header.Get("Content-Type")
	if mime == "" {
		mime = speechMime
	}
	return &core.TTSResult{Audio: audio, MimeType: mime}, nil
}

var audioExtByMime = map[string]string{
	"audio/mpeg":  ".mp3",
	"audio/mp3":   ".mp3",
	"audio/mp4":   ".m4a",
	"audio/m4a":   ".m4a",
	"audio/wav":   ".wav",
	"audio/x-wav": ".wav",
	"audio/webm":  ".webm",
	"audio/ogg":   ".ogg",
	"audio/flac":  ".flac",
}

func audioExt(mime string) string {
	base, _, _ := strings.Cut(mime, ";")
	if ext, ok := audioExtByMime[strings.ToLower(strings.TrimSpace(base))]; ok {
		return ext
	}
	return ".mp3"
}
