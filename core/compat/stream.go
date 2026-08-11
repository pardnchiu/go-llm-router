package compat

import (
	"context"

	"github.com/pardnchiu/go-llm-router/core"
)

const label = "compat"

func (a *Agent) SendStream(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning core.Reasoning, mode core.Mode) (<-chan core.StreamEvent, error) {
	body := a.buildBody(messages, tools)
	body["stream"] = true
	body["stream_options"] = map[string]any{"include_usage": true}

	resp, err := core.OpenStream(ctx, a.httpClient, a.chatAPI(), a.headers(), body, label)
	if err != nil {
		return nil, err
	}
	return core.StreamChat(resp, label), nil
}
