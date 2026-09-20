package openai

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	imageAPI          = "https://api.openai.com/v1/images/generations"
	imageEditAPI      = "https://api.openai.com/v1/images/edits"
	imageLabel        = "openai image"
	imageFormat       = "png"
	ImageInstructions = "You are an image generation assistant. Use the image_generation tool to produce exactly one image matching the user's prompt. Do not respond with text."
)

type imageEvent struct {
	Type string `json:"type"`
	Item *struct {
		Type          string `json:"type"`
		Result        string `json:"result"`
		RevisedPrompt string `json:"revised_prompt"`
		OutputFormat  string `json:"output_format"`
	} `json:"item,omitempty"`
	Error *struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type imageResponse struct {
	Data []struct {
		B64JSON       string `json:"b64_json"`
		RevisedPrompt string `json:"revised_prompt"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func ImageInput(prompt string, opts core.ImageOptions) []map[string]any {
	content := []map[string]any{{"type": "input_text", "text": prompt}}
	if opts.RefImageB64 != "" {
		content = append(content, map[string]any{
			"type":      "input_image",
			"image_url": core.DataURI(opts.RefMime, opts.RefImageB64),
		})
	}
	return []map[string]any{{"role": "user", "content": content}}
}

func imageBody(model, prompt string, opts core.ImageOptions) map[string]any {
	body := map[string]any{
		"model":         model,
		"prompt":        prompt,
		"n":             1,
		"output_format": imageFormat,
	}
	if size := core.ImagePixelSize(opts); size != "" {
		body["size"] = size
	}
	if opts.Quality != "" {
		body["quality"] = opts.Quality
	}
	return body
}

func (a *Agent) generate(ctx context.Context, prompt string, opts core.ImageOptions) (*imageResponse, error) {
	headers := map[string]string{"Authorization": "Bearer " + a.apiKey}
	result, code, err := go_pkg_http.POST[imageResponse](ctx, a.httpClient, imageAPI, headers, imageBody(a.model, prompt, opts), "json")
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: POST: %w", err)
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("%s: http %d", imageLabel, code)
	}
	return &result, nil
}

func (a *Agent) edit(ctx context.Context, prompt string, opts core.ImageOptions) (*imageResponse, error) {
	raw, err := base64.StdEncoding.DecodeString(opts.RefImageB64)
	if err != nil {
		return nil, fmt.Errorf("base64.Decode: %w", err)
	}

	mime := opts.RefMime
	if mime == "" {
		mime = "image/png"
	}
	body := imageBody(a.model, prompt, opts)
	body["image"] = go_pkg_http.File{
		Name:        "reference" + imageExt(opts.RefMime),
		ContentType: mime,
		Data:        raw,
	}

	headers := map[string]string{"Authorization": "Bearer " + a.apiKey}
	result, code, err := go_pkg_http.POST[imageResponse](ctx, a.httpClient, imageEditAPI, headers, body, "multipart")
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: POST: %w", err)
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("%s: http %d", imageLabel, code)
	}
	return &result, nil
}

func (a *Agent) GenerateImage(ctx context.Context, prompt string, opts core.ImageOptions) (*core.ImageResult, error) {
	var out *imageResponse
	var err error
	if opts.RefImageB64 != "" {
		out, err = a.edit(ctx, prompt, opts)
	} else {
		out, err = a.generate(ctx, prompt, opts)
	}
	if err != nil {
		return nil, err
	}

	if out.Error != nil {
		return nil, fmt.Errorf("%s: %s", imageLabel, out.Error.Message)
	}
	if len(out.Data) == 0 || out.Data[0].B64JSON == "" {
		return nil, fmt.Errorf("%s: no image in response", imageLabel)
	}
	return &core.ImageResult{
		B64:      out.Data[0].B64JSON,
		MimeType: "image/" + imageFormat,
		Revised:  out.Data[0].RevisedPrompt,
	}, nil
}

var imageExtByMime = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/jpg":  ".jpg",
	"image/webp": ".webp",
}

func imageExt(mime string) string {
	base, _, _ := strings.Cut(mime, ";")
	if ext, ok := imageExtByMime[strings.ToLower(strings.TrimSpace(base))]; ok {
		return ext
	}
	return ".png"
}

func ReadImageStream(body io.Reader) (*core.ImageResult, error) {
	var result *core.ImageResult
	var streamErr error

	readErr := core.ScanSSE(bufio.NewReader(io.LimitReader(body, core.StreamBodyLimit)), func(_, data string) bool {
		if strings.TrimSpace(data) == "[DONE]" {
			return false
		}

		var ev imageEvent
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			return true
		}
		if ev.Error != nil {
			streamErr = fmt.Errorf("upstream %s: %s", ev.Error.Code, ev.Error.Message)
			return false
		}
		if ev.Type != "response.output_item.done" || ev.Item == nil || ev.Item.Type != "image_generation_call" {
			return true
		}
		if ev.Item.Result == "" {
			streamErr = fmt.Errorf("image_generation_call missing result")
			return false
		}

		format := ev.Item.OutputFormat
		if format == "" {
			format = "png"
		}
		result = &core.ImageResult{
			B64:      ev.Item.Result,
			MimeType: "image/" + format,
			Revised:  ev.Item.RevisedPrompt,
		}
		return false
	})

	if readErr != nil {
		return nil, fmt.Errorf("image stream read: %w", readErr)
	}
	if streamErr != nil {
		return nil, streamErr
	}
	if result == nil {
		return nil, fmt.Errorf("no image_generation_call event in response")
	}
	return result, nil
}
