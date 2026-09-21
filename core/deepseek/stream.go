package deepseek

import (
	"context"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

const label = "deepseek"

func (a *Agent) SendStream(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (<-chan llmrouter.StreamEvent, error) {
	body := a.buildBody(messages, tools)
	body["stream"] = true
	body["stream_options"] = map[string]any{"include_usage": true}

	resp, err := llmrouter.OpenStream(ctx, a.httpClient, chatAPI, a.headers(), body, label)
	if err != nil {
		return nil, err
	}
	return llmrouter.StreamChat(resp, label), nil
}
