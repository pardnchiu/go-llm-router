package openrouter

import (
	"context"
	"fmt"
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	chatAPI = "https://openrouter.ai/api/v1/chat/completions"
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
		"model":       a.model,
		"messages":    merged,
		"temperature": 0.2,
		"tools":       tools,
	}
	if effort, ok := a.effort(reasoning); ok {
		body["reasoning"] = map[string]any{"effort": effort}
	}
	if fast {
		body["service_tier"] = "priority"
	}
	return body
}

func (a *Agent) headers() map[string]string {
	return map[string]string{
		"Authorization":      "Bearer " + a.apiKey,
		"Content-Type":       "application/json",
		"HTTP-Referer":       "https://agenvoy.com",
		"X-OpenRouter-Title": "Agenvoy",
	}
}

func (a *Agent) Send(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (*llmrouter.Output, int, error) {
	fast := mode == llmrouter.ModeFast && llmrouter.SupportFast("openrouter", a.model)
	result, code, err := go_pkg_http.POST[orOutput](ctx, a.httpClient, chatAPI, a.headers(), a.buildBody(messages, tools, reasoning, fast), "json")
	if err != nil {
		return nil, code, err
	}
	if result.Error != nil {
		return nil, code, fmt.Errorf("%s: %s", label, result.Error.Message)
	}
	if fast {
		llmrouter.WarnFastDowngrade("openrouter", a.model, result.ServiceTier)
	}

	out := result.toOutput()
	return out, code, nil
}

type orOutput struct {
	Choices []struct {
		Message struct {
			Role             string `json:"role"`
			Content          any    `json:"content"`
			Reasoning        string `json:"reasoning"`
			ReasoningDetails []struct {
				Type    string `json:"type"`
				Text    string `json:"text"`
				Summary string `json:"summary"`
			} `json:"reasoning_details"`
			ToolCalls []llmrouter.ToolCall `json:"tool_calls"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage       llmrouter.Usage `json:"usage"`
	ServiceTier string          `json:"service_tier"`
	Error       *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (o *orOutput) toOutput() *llmrouter.Output {
	out := &llmrouter.Output{Usage: o.Usage}
	for _, c := range o.Choices {
		reasoning := c.Message.Reasoning
		if reasoning == "" {
			var sb strings.Builder
			for _, d := range c.Message.ReasoningDetails {
				seg := d.Text
				if seg == "" {
					seg = d.Summary
				}
				if seg == "" {
					continue
				}
				if sb.Len() > 0 {
					sb.WriteByte('\n')
				}
				sb.WriteString(seg)
			}
			reasoning = sb.String()
		}
		out.Choices = append(out.Choices, llmrouter.OutputChoices{
			Message: llmrouter.Message{
				Role:             c.Message.Role,
				Content:          c.Message.Content,
				ReasoningContent: reasoning,
				ToolCalls:        c.Message.ToolCalls,
			},
			FinishReason: c.FinishReason,
		})
	}
	return out
}
