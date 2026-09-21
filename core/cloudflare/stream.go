package cloudflare

import (
	"context"
	"fmt"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

const label = "cloudflare"

func (a *Agent) SendStream(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (<-chan llmrouter.StreamEvent, error) {
	return nil, fmt.Errorf("%s: %w", label, llmrouter.ErrStreamUnsupported)
}
