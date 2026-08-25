package grok

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	imageModel   = "grok-imagine-image-2.0"
	imageAPI     = "https://api.x.ai/v1/images/generations"
	imageEditAPI = "https://api.x.ai/v1/images/edits"
)

type imageResponse struct {
	Data []struct {
		B64JSON  string `json:"b64_json"`
		MimeType string `json:"mime_type"`
	} `json:"data"`
	Error any `json:"error"`
}

func imageBody(prompt string, opts core.ImageOptions) map[string]any {
	body := map[string]any{
		"model":           imageModel,
		"prompt":          prompt,
		"n":               1,
		"response_format": "b64_json",
	}
	switch strings.ToLower(opts.Size) {
	case "1k":
		body["resolution"] = "1k"
	case "2k", "4k":
		body["resolution"] = "2k"
	}
	if opts.AspectRatio != "" {
		body["aspect_ratio"] = opts.AspectRatio
	}
	if opts.Quality != "" {
		body["quality"] = opts.Quality
	}
	if opts.RefImageB64 != "" {
		body["image"] = map[string]any{"url": core.DataURI(opts.RefMime, opts.RefImageB64)}
	}
	return body
}

func RequestImage(ctx context.Context, client *http.Client, headers map[string]string, prompt string, opts core.ImageOptions) (*core.ImageResult, error) {
	endpoint := imageAPI
	if opts.RefImageB64 != "" {
		endpoint = imageEditAPI
	}

	result, code, err := go_pkg_http.POST[imageResponse](ctx, client, endpoint, headers, imageBody(prompt, opts), "json")
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("xai: http %d", code)
	}
	if len(result.Data) == 0 || result.Data[0].B64JSON == "" {
		return nil, fmt.Errorf("xai: no image in response")
	}

	mime := result.Data[0].MimeType
	if mime == "" {
		mime = "image/jpeg"
	}
	return &core.ImageResult{B64: result.Data[0].B64JSON, MimeType: mime}, nil
}

func (a *Agent) GenerateImage(ctx context.Context, prompt string, opts core.ImageOptions) (*core.ImageResult, error) {
	return RequestImage(ctx, a.httpClient, a.headers(), prompt, opts)
}
