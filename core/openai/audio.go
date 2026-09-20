package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	transcriptionAPI = "https://api.openai.com/v1/audio/transcriptions"
	speechAPI        = "https://api.openai.com/v1/audio/speech"
	defaultVoice     = "alloy"
	defaultFormat    = "wav"
)

func (a *Agent) Transcribe(ctx context.Context, audio []byte, opts core.STTOptions) (*core.STTResult, error) {
	if len(audio) == 0 {
		return nil, fmt.Errorf("openai.Transcribe: audio is empty")
	}

	body := map[string]any{
		"model": a.model,
		"file": go_pkg_http.File{
			Name:        "audio" + audioExt(opts.MimeType),
			ContentType: opts.MimeType,
			Data:        audio,
		},
	}
	if opts.Language != "" {
		body["language"] = opts.Language
	}
	if opts.Prompt != "" {
		body["prompt"] = opts.Prompt
	}

	headers := map[string]string{"Authorization": "Bearer " + a.apiKey}
	result, code, err := go_pkg_http.POST[struct {
		Text string `json:"text"`
	}](ctx, a.httpClient, transcriptionAPI, headers, body, "multipart")
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: POST: %w", err)
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("openai transcriptions: http %d", code)
	}
	return &core.STTResult{Text: strings.TrimSpace(result.Text)}, nil
}

func (a *Agent) Speak(ctx context.Context, text string, opts core.TTSOptions) (*core.TTSResult, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("openai.Speak: text is empty")
	}

	voice := opts.Voice
	if voice == "" {
		voice = defaultVoice
	}
	format := strings.ToLower(strings.TrimSpace(opts.Format))
	if format == "" {
		format = defaultFormat
	}
	payload, err := json.Marshal(map[string]any{
		"model":           a.model,
		"input":           text,
		"voice":           voice,
		"response_format": format,
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
		mime = speechMime(format)
	}
	return &core.TTSResult{Audio: audio, MimeType: mime}, nil
}

func speechMime(format string) string {
	switch format {
	case "mp3":
		return "audio/mpeg"
	case "opus":
		return "audio/ogg"
	case "aac":
		return "audio/aac"
	case "flac":
		return "audio/flac"
	case "pcm":
		return "audio/pcm"
	default:
		return "audio/wav"
	}
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
