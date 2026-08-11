package cloudflare

import (
	"context"
	"fmt"

	"github.com/pardnchiu/go-llm-router/core"
)

const label = "cloudflare"

func (a *Agent) SendStream(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning core.Reasoning, mode core.Mode) (<-chan core.StreamEvent, error) {
	return nil, fmt.Errorf("%s: %w", label, core.ErrStreamUnsupported)
}
