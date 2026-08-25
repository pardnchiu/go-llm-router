package openai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
)

const (
	imageModel        = "gpt-image-2"
	imageLabel        = "openai image"
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

func imageTool(opts core.ImageOptions) map[string]any {
	tool := map[string]any{"type": "image_generation", "model": imageModel}
	if size := core.ImagePixelSize(opts); size != "" {
		tool["size"] = size
	}
	if opts.Quality != "" {
		tool["quality"] = opts.Quality
	}
	return tool
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

func (a *Agent) GenerateImage(ctx context.Context, prompt string, opts core.ImageOptions) (*core.ImageResult, error) {
	body := map[string]any{
		"model":        a.model,
		"instructions": ImageInstructions,
		"input":        ImageInput(prompt, opts),
		"tools":        []map[string]any{imageTool(opts)},
		"store":        false,
		"stream":       true,
	}

	headers := map[string]string{
		"Authorization": "Bearer " + a.apiKey,
		"Content-Type":  "application/json",
	}

	resp, err := core.OpenStream(ctx, a.httpClient, responsesAPI, headers, body, imageLabel)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ReadImageStream(resp.Body)
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
