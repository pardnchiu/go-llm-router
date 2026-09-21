package openaicodex

import (
	"context"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

const label = "codex"

func (a *Agent) SendStream(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (<-chan llmrouter.StreamEvent, error) {
	headers, err := a.headers(ctx)
	if err != nil {
		return nil, err
	}

	resp, err := llmrouter.OpenStream(ctx, a.httpClient, responsesAPI, headers, a.buildBody(messages, tools, reasoning), label)
	if err != nil {
		return nil, err
	}
	return llmrouter.StreamResponses(resp, label), nil
}
