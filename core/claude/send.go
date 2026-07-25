package claude

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	messagesAPI = "https://api.anthropic.com/v1/messages"
)

func (a *Agent) buildRequestBody(messages []core.Message, tools []core.Tool, reasoning string) map[string]any {
	var systemPrompts []map[string]any
	var newMessages []map[string]any

	for _, msg := range messages {
		if msg.Role == "system" {
			if content, ok := msg.Content.(string); ok && content != "" {
				systemPrompts = append(systemPrompts, map[string]any{
					"type": "text",
					"text": content,
				})
			}
			continue
		}

		message := a.convertToMessage(msg)
		newMessages = append(newMessages, message)
	}

	if len(systemPrompts) > 0 {
		systemPrompts[len(systemPrompts)-1]["cache_control"] = map[string]any{"type": "ephemeral"}
	}
	if len(newMessages) > 0 {
		markCacheControl(newMessages[len(newMessages)-1])
	}

	newTools := a.convertToTools(tools)

	thinkingType := core.GetThinkingType("claude", a.model)
	level := core.ClampReasoningLevel(reasoning, core.MaxReasoningLevel("claude", a.model))
	if core.ReasoningDisabled(level) {
		thinkingType = ""
	}

	requestBody := map[string]any{
		"model":      a.model,
		"max_tokens": a.maxOutputTokens(),
		"messages":   newMessages,
		"tools":      newTools,
	}
	if len(systemPrompts) > 0 {
		requestBody["system"] = systemPrompts
	}
	switch thinkingType {
	case "adaptive":
		requestBody["thinking"] = map[string]any{"type": "adaptive"}
		requestBody["output_config"] = map[string]any{"effort": level}
	case "enabled":
		budget := map[string]int{"low": 5000, "medium": 10000, "high": 32000}[level]
		if budget == 0 {
			budget = 10000
		}
		requestBody["thinking"] = map[string]any{
			"type":          "enabled",
			"budget_tokens": budget,
		}
	default:
		requestBody["temperature"] = 0.2
	}

	return requestBody
}

func (a *Agent) Send(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning string) (*core.Output, int, error) {
	requestBody := a.buildRequestBody(messages, tools, reasoning)

	result, code, err := go_pkg_http.POST[Output](ctx, a.httpClient, messagesAPI, map[string]string{
		"x-api-key":         a.apiKey,
		"anthropic-version": "2023-06-01",
		"anthropic-beta":    "prompt-caching-2024-07-31",
		"Content-Type":      "application/json",
	}, requestBody, "json")
	if err != nil {
		return nil, code, err
	}
	if result.Error != nil {
		return nil, code, fmt.Errorf("%s", result.Error.Message)
	}
	out, err := a.convertToOutput(&result)
	if err != nil {
		return nil, code, err
	}
	return out, code, nil
}

func (a *Agent) convertToMessage(message core.Message) map[string]any {
	if message.ToolCallID != "" {
		var toolResultContent any = message.Content
		if parts, ok := message.Content.([]core.ContentPart); ok {
			var blocks []map[string]any
			for _, p := range parts {
				switch p.Type {
				case "text":
					blocks = append(blocks, map[string]any{"type": "text", "text": p.Text})
				case "image_url":
					if p.ImageURL == nil {
						continue
					}
					mediaType, data, ok := parseDataURL(p.ImageURL.URL)
					if !ok {
						continue
					}
					blocks = append(blocks, map[string]any{
						"type": "image",
						"source": map[string]any{
							"type":       "base64",
							"media_type": mediaType,
							"data":       data,
						},
					})
				}
			}
			toolResultContent = blocks
		}
		return map[string]any{
			"role": "user",
			"content": []map[string]any{
				{
					"type":        "tool_result",
					"tool_use_id": message.ToolCallID,
					"content":     toolResultContent,
				},
			},
		}
	}

	if len(message.ToolCalls) > 0 {
		var content []map[string]any
		for _, tool := range message.ToolCalls {
			var input map[string]any
			json.Unmarshal([]byte(tool.Function.Arguments), &input)
			content = append(content, map[string]any{
				"type":  "tool_use",
				"id":    tool.ID,
				"name":  tool.Function.Name,
				"input": input,
			})
		}
		return map[string]any{
			"role":    message.Role,
			"content": content,
		}
	}

	if parts, ok := message.Content.([]core.ContentPart); ok {
		var content []map[string]any
		for _, part := range parts {
			if part.Type == "text" {
				content = append(content, map[string]any{
					"type": "text",
					"text": part.Text,
				})
			} else if part.Type == "image_url" && part.ImageURL != nil {
				mediaType, data, ok := parseDataURL(part.ImageURL.URL)
				if !ok {
					continue
				}
				content = append(content, map[string]any{
					"type": "image",
					"source": map[string]any{
						"type":       "base64",
						"media_type": mediaType,
						"data":       data,
					},
				})
			}
		}
		return map[string]any{
			"role":    message.Role,
			"content": content,
		}
	}

	return map[string]any{
		"role":    message.Role,
		"content": message.Content,
	}
}

func markCacheControl(message map[string]any) {
	content, ok := message["content"].([]map[string]any)
	if !ok {
		text, isStr := message["content"].(string)
		if !isStr {
			return
		}
		content = []map[string]any{{"type": "text", "text": text}}
		message["content"] = content
	}
	if len(content) == 0 {
		return
	}
	content[len(content)-1]["cache_control"] = map[string]any{"type": "ephemeral"}
}

func parseDataURL(url string) (mediaType, data string, ok bool) {
	// data:<mediaType>;base64,<data>
	if len(url) < 5 || url[:5] != "data:" {
		return "", "", false
	}
	rest := url[5:]
	semi := strings.Index(rest, ";base64,")
	if semi < 0 {
		return "", "", false
	}
	return rest[:semi], rest[semi+8:], true
}

func (a *Agent) convertToTools(tools []core.Tool) []map[string]any {
	newTools := make([]map[string]any, len(tools))
	for i, tool := range tools {
		newTools[i] = map[string]any{
			"name":         tool.Function.Name,
			"description":  tool.Function.Description,
			"input_schema": json.RawMessage(tool.Function.Parameters),
		}
	}
	if len(newTools) > 0 {
		newTools[len(newTools)-1]["cache_control"] = map[string]any{"type": "ephemeral"}
	}
	return newTools
}

func (a *Agent) convertToOutput(resp *Output) (*core.Output, error) {
	output := &core.Output{
		Choices: make([]core.OutputChoices, 1),
		Usage: core.Usage{
			Input:       resp.Usage.InputTokens,
			Output:      resp.Usage.OutputTokens,
			CacheCreate: resp.Usage.CacheCreationInputTokens,
			CacheRead:   resp.Usage.CacheReadInputTokens,
		},
	}

	var toolCalls []core.ToolCall
	var textContent strings.Builder
	var reasoning strings.Builder

	for _, item := range resp.Content {
		if item.Type == "text" {
			textContent.WriteString(item.Text)
		} else if item.Type == "thinking" {
			reasoning.WriteString(item.Thinking)
		} else if item.Type == "tool_use" {
			arg := ""
			if item.Input != nil {
				raw, err := json.Marshal(item.Input)
				if err != nil {
					continue
				}
				arg = string(raw)
			}

			toolCall := core.ToolCall{
				ID:   item.ID,
				Type: "function",
			}
			toolCall.Function.Name = item.Name
			toolCall.Function.Arguments = arg
			toolCalls = append(toolCalls, toolCall)
		}
	}

	text := textContent.String()
	if text == "" && len(toolCalls) == 0 &&
		resp.StopReason != "" && resp.StopReason != "end_turn" && resp.StopReason != "stop_sequence" {
		return nil, fmt.Errorf("claude returned no content (stopReason: %s)", resp.StopReason)
	}

	output.Choices[0].Message = core.Message{
		Role:             "assistant",
		Content:          text,
		ReasoningContent: reasoning.String(),
		ToolCalls:        toolCalls,
	}
	output.Choices[0].FinishReason = resp.StopReason

	return output, nil
}
