package grok

import (
	"context"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/xai"
)

const label = "grok"

func (a *Agent) SendStream(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (<-chan llmrouter.StreamEvent, error) {
	fast := mode == llmrouter.ModeFast && llmrouter.SupportFast(label, a.model)

	resp, err := llmrouter.OpenStream(ctx, a.httpClient, xai.ResponsesAPI, a.headers(), a.buildBody(ctx, messages, tools, reasoning, fast), label)
	if err != nil {
		return nil, err
	}
	return llmrouter.StreamResponses(resp, label), nil
}
