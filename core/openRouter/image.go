package openrouter

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const imageAPI = "https://openrouter.ai/api/v1/images"

type imageResponse struct {
	Data []struct {
		B64JSON   string `json:"b64_json"`
		MediaType string `json:"media_type"`
	} `json:"data"`
}

func (a *Agent) GenerateImage(ctx context.Context, prompt string, opts llmrouter.ImageOptions) (*llmrouter.ImageResult, error) {
	body := map[string]any{
		"model":  a.model,
		"prompt": prompt,
	}
	if opts.AspectRatio != "" {
		body["aspect_ratio"] = opts.AspectRatio
	}
	if opts.Size != "" {
		body["resolution"] = strings.ToUpper(opts.Size)
	}
	if opts.Quality != "" {
		body["quality"] = opts.Quality
	}
	if opts.RefImageB64 != "" {
		body["input_references"] = []map[string]any{{
			"type":      "image_url",
			"image_url": map[string]any{"url": llmrouter.DataURI(opts.RefMime, opts.RefImageB64)},
		}}
	}

	result, code, err := go_pkg_http.POST[imageResponse](ctx, a.httpClient, imageAPI, a.headers(), body, "json")
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: POST: %w", err)
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("openrouter image: http %d", code)
	}
	if len(result.Data) == 0 || result.Data[0].B64JSON == "" {
		return nil, fmt.Errorf("openrouter image: no image in response")
	}

	mime := result.Data[0].MediaType
	if mime == "" {
		mime = "image/png"
	}
	return &llmrouter.ImageResult{B64: result.Data[0].B64JSON, MimeType: mime}, nil
}
