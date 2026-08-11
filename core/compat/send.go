package compat

import (
	"context"
	"fmt"

	"github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

func (a *Agent) chatAPI() string {
	return a.baseURL + "/chat/completions"
}

func (a *Agent) headers() map[string]string {
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if a.apiKey != "" {
		headers["Authorization"] = "Bearer " + a.apiKey
	}
	return headers
}

func (a *Agent) buildBody(messages []core.Message, tools []core.Tool) map[string]any {
	return map[string]any{
		"model":       a.model,
		"messages":    messages,
		"temperature": 0.2,
		"tools":       tools,
	}
}

func (a *Agent) Send(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning core.Reasoning, mode core.Mode) (*core.Output, int, error) {
	result, code, err := go_pkg_http.POST[core.Output](ctx, a.httpClient, a.chatAPI(), a.headers(), a.buildBody(messages, tools), "json")
	if err != nil {
		return nil, code, err
	}
	if result.Error != nil {
		return nil, code, fmt.Errorf("%s: %s", label, result.Error.Message)
	}
	return &result, code, nil
}
