package grok

import (
	"context"
	"fmt"
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	chatAPI = "https://api.x.ai/v1/chat/completions"
)

func (a *Agent) buildBody(messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, fast bool) map[string]any {
	var merged []llmrouter.Message
	var systemParts []string
	for _, m := range messages {
		if m.Role == "system" {
			if s, ok := m.Content.(string); ok && s != "" {
				systemParts = append(systemParts, s)
			}
		} else {
			merged = append(merged, m)
		}
	}
	if len(systemParts) > 0 {
		merged = append([]llmrouter.Message{{Role: "system", Content: strings.Join(systemParts, "\n\n")}}, merged...)
	}

	body := map[string]any{
		"model":    a.model,
		"messages": merged,
		"tools":    tools,
	}
	if llmrouter.SupportTemperature("grok", a.model) {
		body["temperature"] = 0.2
	}
	if effort, ok := a.effort(reasoning); ok {
		body["reasoning_effort"] = effort
	}
	if fast {
		body["service_tier"] = "priority"
	}
	return body
}

func (a *Agent) headers() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + a.apiKey,
		"Content-Type":  "application/json",
	}
}

func (a *Agent) Send(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (*llmrouter.Output, int, error) {
	fast := mode == llmrouter.ModeFast && llmrouter.SupportFast("grok", a.model)

	out, code, err := go_pkg_http.POST[llmrouter.Output](ctx, a.httpClient, chatAPI, a.headers(), a.buildBody(messages, tools, reasoning, fast), "json")
	if err != nil {
		return nil, code, err
	}
	if out.Error != nil {
		return nil, code, fmt.Errorf("%s: %s", label, out.Error.Message)
	}
	if fast {
		llmrouter.WarnFastDowngrade("grok", a.model, out.ServiceTier)
	}
	return &out, code, nil
}
