package grok

import (
	"context"
	"fmt"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	chatAPI = "https://api.x.ai/v1/chat/completions"
)

func (a *Agent) Send(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning string) (*core.Output, int, error) {
	var merged []core.Message
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
		merged = append([]core.Message{{Role: "system", Content: strings.Join(systemParts, "\n\n")}}, merged...)
	}

	body := map[string]any{
		"model":    a.model,
		"messages": merged,
		"tools":    tools,
	}
	if core.SupportTemperature("grok", a.model) {
		body["temperature"] = 0.2
	}
	if core.SupportReasoningEffort("grok", a.model) {
		effort := core.ClampReasoningLevel(reasoning, core.MaxReasoningLevel("grok", a.model))
		if !core.ReasoningDisabled(effort) {
			body["reasoning_effort"] = effort
		}
	}

	out, code, err := go_pkg_http.POST[core.Output](ctx, a.httpClient, chatAPI, map[string]string{
		"Authorization": "Bearer " + a.apiKey,
		"Content-Type":  "application/json",
	}, body, "json")
	if err != nil {
		return nil, code, err
	}
	if out.Error != nil {
		return nil, code, fmt.Errorf("%s", out.Error.Message)
	}
	return &out, code, nil
}
