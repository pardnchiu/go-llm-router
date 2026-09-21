package deepseek

import (
	"context"
	"fmt"
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	chatAPI = "https://api.deepseek.com/v1/chat/completions"
)

func (a *Agent) buildBody(messages []llmrouter.Message, tools []llmrouter.Tool) map[string]any {
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

	for i := range merged {
		if merged[i].Role == "assistant" && merged[i].ReasoningContent == "" {
			merged[i].ReasoningContent = "(reasoning omitted)"
		}
	}

	body := map[string]any{
		"model":    a.model,
		"messages": merged,
		"tools":    tools,
	}
	if llmrouter.SupportTemperature("deepseek", a.model) {
		body["temperature"] = 0.2
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
	result, code, err := go_pkg_http.POST[llmrouter.Output](ctx, a.httpClient, chatAPI, a.headers(), a.buildBody(messages, tools), "json")
	if err != nil {
		return nil, code, err
	}
	if result.Error != nil {
		return nil, code, fmt.Errorf("%s: %s", label, result.Error.Message)
	}
	return &result, code, nil
}
