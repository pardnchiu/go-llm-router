package gemini

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const imageModel = "gemini-3.1-flash-image"

type imageResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text       string `json:"text"`
				InlineData *struct {
					MimeType string `json:"mimeType"`
					Data     string `json:"data"`
				} `json:"inlineData"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (a *Agent) GenerateImage(ctx context.Context, prompt string, opts core.ImageOptions) (*core.ImageResult, error) {
	parts := []map[string]any{}
	if opts.RefImageB64 != "" {
		mime := opts.RefMime
		if mime == "" {
			mime = "image/png"
		}
		parts = append(parts, map[string]any{
			"inline_data": map[string]any{"mime_type": mime, "data": opts.RefImageB64},
		})
	}
	parts = append(parts, map[string]any{"text": prompt})

	// * skip quality
	imageConfig := map[string]any{}
	if opts.AspectRatio != "" {
		imageConfig["aspectRatio"] = opts.AspectRatio
	}
	if opts.Size != "" {
		imageConfig["imageSize"] = strings.ToUpper(opts.Size)
	}

	body := map[string]any{
		"contents": []map[string]any{{"role": "user", "parts": parts}},
	}
	if len(imageConfig) > 0 {
		body["generationConfig"] = map[string]any{"imageConfig": imageConfig}
	}

	endpoint := baseAPI + imageModel + ":generateContent"
	result, code, err := go_pkg_http.POST[imageResponse](ctx, a.httpClient, endpoint, a.headers(), body, "json")
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, fmt.Errorf("gemini: %s", result.Error.Message)
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("gemini: http %d", code)
	}

	for _, c := range result.Candidates {
		for _, p := range c.Content.Parts {
			if p.InlineData != nil && p.InlineData.Data != "" {
				return &core.ImageResult{B64: p.InlineData.Data, MimeType: p.InlineData.MimeType}, nil
			}
		}
	}
	return nil, fmt.Errorf("gemini: no inlineData part in response")
}
