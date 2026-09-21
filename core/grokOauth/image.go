package grokoauth

import (
	"context"
	"fmt"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/grok"
)

func (a *Agent) GenerateImage(ctx context.Context, prompt string, opts llmrouter.ImageOptions) (*llmrouter.ImageResult, error) {
	headers, err := a.headers(ctx)
	if err != nil {
		return nil, fmt.Errorf("a.headers: %w", err)
	}
	return grok.RequestImage(ctx, llmrouter.NewHTTPClient(), headers, a.model, prompt, opts)
}
