package ollamacloud

import (
	"context"
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	chatAPI = "https://ollama.com/v1/chat/completions"
)

type response struct {
	Choices []struct {
		Message struct {
			Role      string               `json:"role"`
			Content   string               `json:"content"`
			Reasoning string               `json:"reasoning"`
			ToolCalls []llmrouter.ToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage llmrouter.Usage `json:"usage"`
}

func (a *Agent) buildBody(messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning) map[string]any {
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

	return map[string]any{
		"model":            a.model,
		"messages":         merged,
		"temperature":      0.2,
		"tools":            tools,
		"reasoning_effort": a.effort(reasoning),
	}
}

func (a *Agent) headers() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + a.apiKey,
		"Content-Type":  "application/json",
	}
}

func (a *Agent) Send(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (*llmrouter.Output, int, error) {
	result, code, err := go_pkg_http.POST[response](ctx, a.httpClient, chatAPI, a.headers(), a.buildBody(messages, tools, reasoning), "json")
	if err != nil {
		return nil, code, err
	}

	output := &llmrouter.Output{
		Choices: make([]llmrouter.OutputChoices, 0, len(result.Choices)),
		Usage:   result.Usage,
	}
	for _, c := range result.Choices {
		output.Choices = append(output.Choices, llmrouter.OutputChoices{
			Message: llmrouter.Message{
				Role:             c.Message.Role,
				Content:          c.Message.Content,
				ReasoningContent: c.Message.Reasoning,
				ToolCalls:        c.Message.ToolCalls,
			},
			FinishReason: c.FinishReason,
		})
	}
	return output, code, nil
}
