package typesafe

import (
	"context"
	"fmt"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const systemOneAPI = "https://api.typesafe.ai/v1/systemone"

func (a *Agent) Decide(ctx context.Context, state any, questions map[string]llmrouter.DecisionQuestion) (*llmrouter.DecisionResult, error) {
	body, err := llmrouter.DecisionBody(a.model, state, questions)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Authorization": "Bearer " + a.apiKey,
		"Content-Type":  "application/json",
	}

	result, code, err := go_pkg_http.POST[llmrouter.DecisionResponse](ctx, a.httpClient, systemOneAPI, headers, body, "json")
	if err != nil {
		return nil, fmt.Errorf("[%s] typesafe systemone: %w", llmrouter.HTTPErrorLabel(code, err), err)
	}
	return result.Result("typesafe systemone")
}
