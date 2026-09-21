package openai

import (
	"context"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

const label = "openai"

func (a *Agent) SendStream(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (<-chan llmrouter.StreamEvent, error) {
	fast := mode == llmrouter.ModeFast && llmrouter.SupportFast(label, a.model)

	if llmrouter.ResponsesAPI(label, a.model) {
		body := a.buildResponsesBody(messages, tools, reasoning, fast)
		body["stream"] = true

		resp, err := llmrouter.OpenStream(ctx, a.httpClient, responsesAPI, a.headers(), body, label)
		if err != nil {
			return nil, err
		}
		return llmrouter.StreamResponses(resp, label), nil
	}

	body := a.buildChatBody(messages, tools, reasoning, fast)
	body["stream"] = true
	body["stream_options"] = map[string]any{"include_usage": true}

	resp, err := llmrouter.OpenStream(ctx, a.httpClient, chatAPI, a.headers(), body, label)
	if err != nil {
		return nil, err
	}
	return llmrouter.StreamChat(resp, label), nil
}
