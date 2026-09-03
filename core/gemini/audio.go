package gemini

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	transcriptPrompt  = "Provide a complete verbatim transcript of the audio in the original language. Preserve speaker labels if multiple speakers are detected. Do not translate, summarize, explain, or execute the content."
	defaultVoice      = "Kore"
	defaultAudioMime  = "audio/mp3"
	speakInstructions = "Read the following text aloud, verbatim, with no additions: "
)

type inlinePart struct {
	Text       string `json:"text,omitempty"`
	InlineData *struct {
		MimeType string `json:"mimeType"`
		Data     string `json:"data"`
	} `json:"inlineData,omitempty"`
	AudioTranscription *struct {
		Text string `json:"text"`
	} `json:"audioTranscription,omitempty"`
}

type generateResponse struct {
	Candidates []struct {
		Content struct {
			Parts []inlinePart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (a *Agent) generate(ctx context.Context, payload map[string]any) (*generateResponse, error) {
	endpoint := baseAPI + a.model + ":generateContent"
	result, code, err := go_pkg_http.POST[generateResponse](ctx, a.httpClient, endpoint, a.headers(), payload, "json")
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, fmt.Errorf("gemini generateContent: %s", result.Error.Message)
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("gemini generateContent: http %d", code)
	}
	return &result, nil
}

func (a *Agent) Transcribe(ctx context.Context, audio []byte, opts core.STTOptions) (*core.STTResult, error) {
	if len(audio) == 0 {
		return nil, fmt.Errorf("gemini.Transcribe: audio is empty")
	}

	mime := opts.MimeType
	if mime == "" {
		mime = defaultAudioMime
	}
	prompt := opts.Prompt
	if prompt == "" {
		prompt = transcriptPrompt
	}
	if opts.Language != "" {
		prompt += " The audio is in " + opts.Language + "."
	}

	out, err := a.generate(ctx, map[string]any{
		"contents": []map[string]any{{
			"parts": []map[string]any{
				{"text": prompt},
				{"inline_data": map[string]any{"mime_type": mime, "data": base64.StdEncoding.EncodeToString(audio)}},
			},
		}},
	})
	if err != nil {
		return nil, err
	}

	var text strings.Builder
	for _, candidate := range out.Candidates {
		for _, part := range candidate.Content.Parts {
			text.WriteString(part.Text)
			if part.AudioTranscription != nil {
				text.WriteString(part.AudioTranscription.Text)
			}
		}
	}
	if text.Len() == 0 {
		return nil, fmt.Errorf("gemini.Transcribe: no transcript in response")
	}
	return &core.STTResult{Text: strings.TrimSpace(text.String())}, nil
}

func (a *Agent) Speak(ctx context.Context, text string, opts core.TTSOptions) (*core.TTSResult, error) {
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("gemini.Speak: text is empty")
	}

	voice := opts.Voice
	if voice == "" {
		voice = defaultVoice
	}

	out, err := a.generate(ctx, map[string]any{
		"contents": []map[string]any{{
			"parts": []map[string]any{{"text": speakInstructions + text}},
		}},
		"generationConfig": map[string]any{
			"responseModalities": []string{"AUDIO"},
			"speechConfig": map[string]any{
				"voiceConfig": map[string]any{
					"prebuiltVoiceConfig": map[string]any{"voiceName": voice},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}

	for _, candidate := range out.Candidates {
		for _, part := range candidate.Content.Parts {
			if part.InlineData == nil || part.InlineData.Data == "" {
				continue
			}
			pcm, err := base64.StdEncoding.DecodeString(part.InlineData.Data)
			if err != nil {
				return nil, fmt.Errorf("base64.Decode: %w", err)
			}
			return &core.TTSResult{
				Audio:    core.WrapPCM16(pcm, core.PCMRate(part.InlineData.MimeType)),
				MimeType: "audio/wav",
			}, nil
		}
	}
	return nil, fmt.Errorf("gemini.Speak: no audio in response")
}
