package cloudflare

import (
	"context"
	"fmt"

	"github.com/pardnchiu/go-llm-router/core"
)

func (a *Agent) SendStream(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning core.Reasoning, mode core.Mode) (<-chan core.StreamEvent, error) {
	return nil, fmt.Errorf("cloudflare: %w", core.ErrStreamUnsupported)
}
