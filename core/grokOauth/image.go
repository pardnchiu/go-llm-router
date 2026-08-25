package grokoauth

import (
	"context"
	"fmt"

	"github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/grok"
)

func (a *Agent) GenerateImage(ctx context.Context, prompt string, opts core.ImageOptions) (*core.ImageResult, error) {
	headers, err := a.headers(ctx)
	if err != nil {
		return nil, fmt.Errorf("a.headers: %w", err)
	}
	return grok.RequestImage(ctx, core.NewHTTPClient(), headers, prompt, opts)
}
