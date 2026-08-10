package openrouter

import (
	"context"

	"github.com/pardnchiu/go-llm-router/core"
)

const label = "openrouter"

func (a *Agent) SendStream(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning core.Reasoning, mode core.Mode) (<-chan core.StreamEvent, error) {
	fast := mode == core.ModeFast && core.SupportFast(label, a.model)
	body := a.buildBody(messages, tools, reasoning, fast)
	body["stream"] = true
	body["stream_options"] = map[string]any{"include_usage": true}

	resp, _, err := core.OpenStream(ctx, a.httpClient, chatAPI, a.headers(), body, label)
	if err != nil {
		return nil, err
	}
	return core.StreamChat(resp, label), nil
}
