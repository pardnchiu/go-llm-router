package copilot

import (
	"context"
	"fmt"

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

func (a *Agent) buildResponsesBody(messages []core.Message, tools []core.Tool, reasoning string) map[string]any {
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

	effort := core.ClampReasoningLevel(reasoning, core.MaxReasoningLevel("copilot", a.model))
	body := map[string]any{
		"model":        a.model,
		"input":        copilotResponse.ConvertInput(nonSystem),
		"tools":        copilotResponse.ConvertTools(tools),
		"instructions": instructions,
		"store":        false,
	}
	if !core.ReasoningDisabled(effort) {
		body["reasoning"] = map[string]any{"effort": effort, "summary": "auto"}
	}
	return body
}

func (a *Agent) buildChatBody(messages []core.Message, tools []core.Tool, reasoning string) map[string]any {
	body := map[string]any{
		"model":    a.model,
		"messages": messages,
		"tools":    tools,
	}
	if core.SupportTemperature("copilot", a.model) {
		body["temperature"] = 0.2
	}
	if core.SupportReasoningEffort("copilot", a.model) {
		effort := core.ClampReasoningLevel(reasoning, core.MaxReasoningLevel("copilot", a.model))
		if !core.ReasoningDisabled(effort) {
			body["reasoning_effort"] = effort
		}
	}
	return body
}

func (a *Agent) Send(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning string) (*core.Output, int, error) {
	headers, err := a.headers(ctx)
	if err != nil {
		return nil, 0, err
	}

	if core.ResponsesAPI("copilot", a.model) {
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
