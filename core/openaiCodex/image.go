package openaicodex

import (
	"context"
	"fmt"

	"github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/openai"
)

const imageLabel = "codex image"

func (a *Agent) GenerateImage(ctx context.Context, prompt string, opts core.ImageOptions) (*core.ImageResult, error) {
	auth, err := a.authHeader(ctx)
	if err != nil {
		return nil, fmt.Errorf("a.authHeader: %w", err)
	}

	headers := map[string]string{
		"Authorization": auth,
		"Content-Type":  "application/json",
	}
	if a.token != nil && a.token.AccountID != "" {
		headers["ChatGPT-Account-Id"] = a.token.AccountID
	}

	body := map[string]any{
		"model":        a.model,
		"instructions": openai.ImageInstructions,
		"input":        openai.ImageInput(prompt, opts),
		"tools":        []map[string]any{{"type": "image_generation"}},
		"store":        false,
		"stream":       true,
	}

	resp, err := core.OpenStream(ctx, a.httpClient, responsesAPI, headers, body, imageLabel)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return openai.ReadImageStream(resp.Body)
}
