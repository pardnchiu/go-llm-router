package copilot

import (
	"context"
	"fmt"
	"slices"

	"github.com/pardnchiu/go-llm-router/core"
	copilotResponse "github.com/pardnchiu/go-llm-router/core/copilot/response"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	chatAPI      = "https://api.githubcopilot.com/chat/completions"
	responsesAPI = "https://api.githubcopilot.com/responses"
)

func (a *Agent) headers(ctx context.Context) (map[string]string, error) {
	auth, err := a.authHeader(ctx)
	if err != nil {
		return nil, fmt.Errorf("a.authHeader: %w", err)
	}
	return map[string]string{
		"Authorization":  auth,
		"Editor-Version": "vscode/1.95.0",
	}, nil
}

func (a *Agent) buildResponsesBody(messages []core.Message, tools []core.Tool, reasoning core.Reasoning) map[string]any {
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
	return body
}

func (a *Agent) buildChatBody(messages []core.Message, tools []core.Tool, reasoning core.Reasoning) map[string]any {
	body := map[string]any{
		"model":    a.model,
		"messages": messages,
		"tools":    tools,
	}
	if core.SupportTemperature("copilot", a.model) {
		body["temperature"] = 0.2
	}
	if effort, ok := a.effort(reasoning); ok {
		body["reasoning_effort"] = effort
	}
	return body
}

func (a *Agent) useResponses() bool {
	if len(a.endpoints) > 0 {
		hasChat := slices.Contains(a.endpoints, "/chat/completions")
		hasResponses := slices.Contains(a.endpoints, "/responses")
		if hasResponses && !hasChat {
			return true
		}
		if hasChat && !hasResponses {
			return false
		}
	}
	return core.ResponsesAPI("copilot", a.model)
}

func (a *Agent) Send(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning core.Reasoning) (*core.Output, int, error) {
	headers, err := a.headers(ctx)
	if err != nil {
		return nil, 0, err
	}

	if a.useResponses() {
		body := a.buildResponsesBody(messages, tools, reasoning)

		result, code, err := go_pkg_http.POST[copilotResponse.Output](ctx, a.httpClient, responsesAPI, headers, body, "json")
		if err != nil {
			return nil, code, err
		}
		if result.Error != nil {
			return nil, code, fmt.Errorf("%s", result.Error.Message)
		}

		out := copilotResponse.ConvertOutput(result)
		return &out, code, nil
	}

	body := a.buildChatBody(messages, tools, reasoning)

	result, code, err := go_pkg_http.POST[core.Output](ctx, a.httpClient, chatAPI, headers, body, "json")
	if err != nil {
		return nil, code, err
	}
	if result.Error != nil {
		return nil, code, fmt.Errorf("%s", result.Error.Message)
	}
	return &result, code, nil
}
