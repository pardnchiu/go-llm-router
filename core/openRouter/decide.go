package openrouter

import (
	"context"
	"fmt"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const decisionsAPI = "https://openrouter.ai/api/alpha/decisions"

func (a *Agent) Decide(ctx context.Context, state any, questions map[string]llmrouter.DecisionQuestion) (*llmrouter.DecisionResult, error) {
	body, err := llmrouter.DecisionBody(a.model, state, questions)
	if err != nil {
		return nil, err
	}

	result, code, err := go_pkg_http.POST[llmrouter.DecisionResponse](ctx, a.httpClient, decisionsAPI, a.headers(), body, "json")
	if err != nil {
		return nil, fmt.Errorf("[%s] openrouter decisions: %w", llmrouter.HTTPErrorLabel(code, err), err)
	}
	return result.Result("openrouter decisions")
}
