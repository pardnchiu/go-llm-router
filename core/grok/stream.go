package grok

import (
	"context"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

const label = "grok"

func (a *Agent) SendStream(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (<-chan llmrouter.StreamEvent, error) {
	fast := mode == llmrouter.ModeFast && llmrouter.SupportFast(label, a.model)
	body := a.buildBody(messages, tools, reasoning, fast)
	body["stream"] = true
	body["stream_options"] = map[string]any{"include_usage": true}

	resp, err := llmrouter.OpenStream(ctx, a.httpClient, chatAPI, a.headers(), body, label)
	if err != nil {
		return nil, err
	}
	return llmrouter.StreamChat(resp, label), nil
}
