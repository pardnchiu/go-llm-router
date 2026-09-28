package grokoauth

import (
	"context"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/xai"
)

const label = "grok-oauth"

func (a *Agent) SendStream(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (<-chan llmrouter.StreamEvent, error) {
	headers, err := a.headers(ctx)
	if err != nil {
		return nil, err
	}
	fast := mode == llmrouter.ModeFast && llmrouter.SupportFast("grok", a.model)

	resp, err := llmrouter.OpenStream(ctx, a.httpClient, xai.ResponsesAPI, headers, a.buildBody(ctx, messages, tools, reasoning, fast), label)
	if err != nil {
		return nil, err
	}
	return llmrouter.StreamResponses(resp, label), nil
}
