package mistral

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	chatAPI = "https://api.mistral.ai/v1/chat/completions"
)

type response struct {
	Choices []struct {
		Message struct {
			Role      string          `json:"role"`
			Content   json.RawMessage `json:"content"`
			ToolCalls []core.ToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage core.Usage `json:"usage"`
}

type contentPart struct {
	Type     string `json:"type"`
	Text     string `json:"text"`
	Thinking []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"thinking"`
}

func splitContent(raw json.RawMessage) (text, reasoning string) {
	if len(raw) == 0 || string(raw) == "null" {
		return "", ""
	}

	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s, ""
	}

	var parts []contentPart
	if json.Unmarshal(raw, &parts) != nil {
		return "", ""
	}

	var textBuilder, reasoningBuilder strings.Builder
	for _, p := range parts {
		switch p.Type {
		case "text":
			textBuilder.WriteString(p.Text)
		case "thinking":
			for _, t := range p.Thinking {
				reasoningBuilder.WriteString(t.Text)
			}
		}
	}
	return textBuilder.String(), reasoningBuilder.String()
}

func (a *Agent) buildBody(messages []core.Message, tools []core.Tool, reasoning core.Reasoning) map[string]any {
	cleaned := make([]core.Message, len(messages))
	copy(cleaned, messages)
	for i := range cleaned {
		cleaned[i].ReasoningContent = ""
	}

	body := map[string]any{
		"model":       a.model,
		"messages":    cleaned,
		"temperature": 0.2,
		"tools":       tools,
	}
	if effort, ok := a.effort(reasoning); ok {
		body["reasoning_effort"] = effort
	}
	return body
}

func (a *Agent) headers() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + a.apiKey,
		"Content-Type":  "application/json",
	}
}

func (a *Agent) Send(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning core.Reasoning, mode core.Mode) (*core.Output, int, error) {
	result, code, err := go_pkg_http.POST[response](ctx, a.httpClient, chatAPI, a.headers(), a.buildBody(messages, tools, reasoning), "json")
	if err != nil {
		return nil, code, err
	}

	output := &core.Output{
		Choices: make([]core.OutputChoices, 0, len(result.Choices)),
		Usage:   result.Usage,
	}
	for _, c := range result.Choices {
		text, reasoningText := splitContent(c.Message.Content)
		output.Choices = append(output.Choices, core.OutputChoices{
			Message: core.Message{
				Role:             c.Message.Role,
				Content:          text,
				ReasoningContent: reasoningText,
				ToolCalls:        c.Message.ToolCalls,
			},
			FinishReason: c.FinishReason,
		})
	}
	return output, code, nil
}
