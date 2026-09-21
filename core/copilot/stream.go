package copilot

import (
	"context"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

const label = "copilot"

func (a *Agent) SendStream(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (<-chan llmrouter.StreamEvent, error) {
	headers, err := a.headers(ctx)
	if err != nil {
		return nil, err
	}

	if a.useResponses(ctx) {
		body := a.buildResponsesBody(messages, tools, reasoning)
		body["stream"] = true

		resp, err := llmrouter.OpenStream(ctx, a.httpClient, responsesAPI, headers, body, label)
		if err != nil {
			return nil, err
		}
		return llmrouter.StreamResponses(resp, label), nil
	}

	body := a.buildChatBody(messages, tools, reasoning)
	body["stream"] = true
	body["stream_options"] = map[string]any{"include_usage": true}

	resp, err := llmrouter.OpenStream(ctx, a.httpClient, chatAPI, headers, body, label)
	if err != nil {
		return nil, err
	}
	return llmrouter.StreamChat(resp, label), nil
}
