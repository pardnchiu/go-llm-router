package grokoauth

import (
	"context"

	"github.com/pardnchiu/go-llm-router/core"
)

const label = "grok-oauth"

func (a *Agent) SendStream(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning core.Reasoning, mode core.Mode) (<-chan core.StreamEvent, error) {
	headers, err := a.headers(ctx)
	if err != nil {
		return nil, err
	}
	fast := mode == core.ModeFast && core.SupportFast("grok", a.model)

	resp, _, err := core.OpenStream(ctx, a.httpClient, responsesAPI, headers, a.buildBody(messages, tools, reasoning, fast), label)
	if err != nil {
		return nil, err
	}
	return core.StreamResponses(resp, label), nil
}
