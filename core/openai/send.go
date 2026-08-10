package openai

import (
	"context"
	"fmt"

	"github.com/pardnchiu/go-llm-router/core"
	copilotResponse "github.com/pardnchiu/go-llm-router/core/copilot/response"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	chatAPI      = "https://api.openai.com/v1/chat/completions"
	responsesAPI = "https://api.openai.com/v1/responses"
)

func (a *Agent) headers() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + a.apiKey,
		"Content-Type":  "application/json",
	}
}

func (a *Agent) buildResponsesBody(messages []core.Message, tools []core.Tool, reasoning core.Reasoning, fast bool) map[string]any {
	var instructions string
	nonSystem := make([]core.Message, 0, len(messages))
	for _, m := range messages {
		if m.Role == "system" {
			if s, ok := m.Content.(string); ok {
				if instructions != "" {
					instructions += "\n"
				}
				instructions += s
			}
			continue
		}
		nonSystem = append(nonSystem, m)
	}

	body := map[string]any{
		"model":        a.model,
		"input":        copilotResponse.ConvertInput(nonSystem),
		"tools":        copilotResponse.ConvertTools(tools),
		"instructions": instructions,
		"store":        false,
	}
	if effort, ok := a.effort(reasoning); ok {
		body["reasoning"] = map[string]any{"effort": effort, "summary": "auto"}
	}
	if fast {
		body["service_tier"] = "priority"
	}
	return body
}

func (a *Agent) buildChatBody(messages []core.Message, tools []core.Tool, reasoning core.Reasoning, fast bool) map[string]any {
	body := map[string]any{
		"model":    a.model,
		"messages": messages,
		"tools":    tools,
	}
	if core.SupportTemperature("openai", a.model) {
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

func (a *Agent) Send(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning core.Reasoning, mode core.Mode) (*core.Output, int, error) {
	fast := mode == core.ModeFast && core.SupportFast("openai", a.model)

	if core.ResponsesAPI("openai", a.model) {
		result, code, err := go_pkg_http.POST[copilotResponse.Output](ctx, a.httpClient, responsesAPI, a.headers(), a.buildResponsesBody(messages, tools, reasoning, fast), "json")
		if err != nil {
			return nil, code, err
		}
		if result.Error != nil {
			return nil, code, fmt.Errorf("%s", result.Error.Message)
		}
		if fast {
			core.WarnFastDowngrade("openai", a.model, result.ServiceTier)
		}

		out := copilotResponse.ConvertOutput(result)
		return &out, code, nil
	}

	result, code, err := go_pkg_http.POST[core.Output](ctx, a.httpClient, chatAPI, a.headers(), a.buildChatBody(messages, tools, reasoning, fast), "json")
	if err != nil {
		return nil, code, err
	}
	if result.Error != nil {
		return nil, code, fmt.Errorf("http.POST: %s", result.Error.Message)
	}
	if fast {
		core.WarnFastDowngrade("openai", a.model, result.ServiceTier)
	}

	return &result, code, nil
}
