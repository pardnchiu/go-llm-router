package copilot

import (
	"context"

	"github.com/pardnchiu/go-llm-router/core"
)

const label = "copilot"

func (a *Agent) SendStream(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning core.Reasoning, mode core.Mode) (<-chan core.StreamEvent, error) {
	headers, err := a.headers(ctx)
	if err != nil {
		return nil, err
	}

	if a.useResponses(ctx) {
		body := a.buildResponsesBody(messages, tools, reasoning)
		body["stream"] = true

		resp, err := core.OpenStream(ctx, a.httpClient, responsesAPI, headers, body, label)
		if err != nil {
			return nil, err
		}
		return core.StreamResponses(resp, label), nil
	}

	body := a.buildChatBody(messages, tools, reasoning)
	body["stream"] = true
	body["stream_options"] = map[string]any{"include_usage": true}

	resp, err := core.OpenStream(ctx, a.httpClient, chatAPI, headers, body, label)
	if err != nil {
		return nil, err
	}
	return core.StreamChat(resp, label), nil
}
