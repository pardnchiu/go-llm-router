package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	baseAPI = "https://generativelanguage.googleapis.com/v1beta/models/"
)

func (a *Agent) Send(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning string) (*core.Output, int, error) {
	messages = rewriteSyntheticActivations(messages)

	var systemParts []string
	var newMessages []Content

	for _, msg := range messages {
		if msg.Role == "system" {
			if content, ok := msg.Content.(string); ok && content != "" {
				systemParts = append(systemParts, content)
			}
			continue
		}

		message := a.convertToContent(msg)
		newMessages = append(newMessages, message)
	}

	systemPrompt := strings.Join(systemParts, "\n\n")
	newTools := a.convertToTools(tools)
	apiURL := fmt.Sprintf("%s%s:generateContent", baseAPI, a.model)

	cachedName, sendMessages := a.applyCache(ctx, systemPrompt, newMessages, newTools)
	requestBody := a.generateRequestBody(sendMessages, systemPrompt, newTools, cachedName, reasoning)

	result, code, err := go_pkg_http.POST[Output](ctx, a.httpClient, apiURL, map[string]string{
		"Content-Type":   "application/json",
		"x-goog-api-key": a.apiKey,
	}, requestBody, "json")
	if err != nil {
		return nil, code, err
	}

	out, err := a.convertToOutput(&result)
	if err != nil {
		return nil, code, err
	}
	return out, code, nil
}

func rewriteSyntheticActivations(messages []core.Message) []core.Message {
	out := make([]core.Message, 0, len(messages))
	for i := 0; i < len(messages); i++ {
		msg := messages[i]
		if msg.Role == "assistant" && len(msg.ToolCalls) == 1 {
			tc := msg.ToolCalls[0]
			if tc.Function.Name == "run_skill" && tc.ThoughtSignature == "" && i+1 < len(messages) {
				next := messages[i+1]
				if next.Role == "tool" && next.ToolCallID == tc.ID {
					activation, _ := next.Content.(string)
					out = append(out, core.Message{
						Role:    "user",
						Content: activation,
					})
					i++
					continue
				}
			}
		}
		out = append(out, msg)
	}
	return out
}

func (a *Agent) convertToContent(message core.Message) Content {
	content := Content{}
	if message.ToolCallID != "" {
		content.Role = "user"
		data := map[string]any{}
		if contentStr, ok := message.Content.(string); ok {
			data["result"] = contentStr
		}
		content.Parts = []Part{
			{
				FunctionResponse: &FunctionResponse{
					Name:     message.ToolCallID,
					Response: data,
				},
			},
		}
		if parts, ok := message.Content.([]core.ContentPart); ok {
			for _, p := range parts {
				if p.Type == "image_url" && p.ImageURL != nil {
					url := p.ImageURL.URL
					if strings.HasPrefix(url, "data:") {
						if semi := strings.Index(url, ";base64,"); semi != -1 {
							mimeType := url[5:semi]
							b64 := url[semi+8:]
							content.Parts = append(content.Parts, Part{
								InlineData: &InlineData{MimeType: mimeType, Data: b64},
							})
						}
					}
				}
			}
		}
		return content
	}

	role := message.Role
	if role == "assistant" {
		role = "model"
	}
	// * gemini contents only accepts user / model; empty role falls back to user
	if role == "" {
		role = "user"
	}
	content.Role = role

	if len(message.ToolCalls) > 0 {
		for _, tool := range message.ToolCalls {
			var args map[string]any
			json.Unmarshal([]byte(tool.Function.Arguments), &args)
			content.Parts = append(content.Parts, Part{
				ThoughtSignature: tool.ThoughtSignature,
				FunctionCall: &FunctionCall{
					Name: tool.Function.Name,
					Args: args,
				},
			})
		}
		return content
	}

	switch v := message.Content.(type) {
	case string:
		content.Parts = []Part{{Text: v}}
	case []core.ContentPart:
		for _, p := range v {
			switch p.Type {
			case "text":
				content.Parts = append(content.Parts, Part{Text: p.Text})
			case "image_url":
				if p.ImageURL == nil {
					continue
				}
				// * to inlineData
				url := p.ImageURL.URL
				if strings.HasPrefix(url, "data:") {
					if semi := strings.Index(url, ";base64,"); semi != -1 {
						mimeType := url[5:semi]
						b64 := url[semi+8:]
						content.Parts = append(content.Parts, Part{
							InlineData: &InlineData{MimeType: mimeType, Data: b64},
						})
					}
				}
			}
		}
	}

	return content
}

func (a *Agent) convertToTools(tools []core.Tool) []map[string]any {
	newTools := make([]map[string]any, len(tools))
	for i, tool := range tools {
		var params map[string]any
		json.Unmarshal(tool.Function.Parameters, &params)
		sanitizeSchema(params)

		newTools[i] = map[string]any{
			"name":        tool.Function.Name,
			"description": tool.Function.Description,
			"parameters":  params,
		}
	}
	return newTools
}

var geminiUnsupportedKeys = map[string]struct{}{
	"$schema":              {},
	"$id":                  {},
	"$ref":                 {},
	"$defs":                {},
	"$comment":             {},
	"definitions":          {},
	"additionalProperties": {},
	"patternProperties":    {},
}

func sanitizeSchema(m map[string]any) {
	for key := range geminiUnsupportedKeys {
		delete(m, key)
	}

	if list, ok := m["enum"].([]any); ok {
		for i, v := range list {
			if _, ok := v.(string); !ok {
				list[i] = fmt.Sprintf("%v", v)
			}
		}
	}

	for _, v := range m {
		switch child := v.(type) {
		case map[string]any:
			sanitizeSchema(child)
		case []any:
			for _, item := range child {
				if obj, ok := item.(map[string]any); ok {
					sanitizeSchema(obj)
				}
			}
		}
	}
}

func (a *Agent) generateRequestBody(messages []Content, prompt string, newTools []map[string]any, cachedContent string, reasoning string) map[string]any {
	thinkingConfig := core.GetThinkingConfig("gemini", a.model)
	level := core.ClampReasoningLevel(reasoning, core.MaxReasoningLevel("gemini", a.model))
	level = core.FloorReasoningLevel(level, core.MinReasoningLevel("gemini", a.model))

	generationConfig := map[string]any{}
	switch {
	case thinkingConfig == "level":
		generationConfig["thinkingConfig"] = map[string]any{
			"thinkingLevel":   level,
			"includeThoughts": true,
		}
	case thinkingConfig == "budget":
		generationConfig["temperature"] = 0.2
		budget := core.ThinkingBudget(a.model, level)
		thinking := map[string]any{"thinkingBudget": budget}
		if budget > 0 {
			thinking["includeThoughts"] = true
		}
		generationConfig["thinkingConfig"] = thinking
	default:
		generationConfig["temperature"] = 0.2
	}
	body := map[string]any{
		"contents":         messages,
		"generationConfig": generationConfig,
	}

	if cachedContent != "" {
		body["cachedContent"] = cachedContent
		return body
	}

	if prompt != "" {
		body["systemInstruction"] = map[string]any{
			"parts": []map[string]any{
				{"text": prompt},
			},
		}
	}

	if len(newTools) > 0 {
		body["tools"] = []map[string]any{
			{"functionDeclarations": newTools},
		}
	}
	return body
}

func (a *Agent) convertToOutput(resp *Output) (*core.Output, error) {
	output := &core.Output{
		Choices: make([]core.OutputChoices, 1),
	}

	if resp.UsageMetadata != nil {
		output.Usage = core.Usage{
			Input:     resp.UsageMetadata.PromptTokenCount - resp.UsageMetadata.CachedContentTokenCount,
			Output:    resp.UsageMetadata.CandidatesTokenCount,
			CacheRead: resp.UsageMetadata.CachedContentTokenCount,
		}
	}

	if len(resp.Candidates) == 0 {
		reason := "no candidates"
		if resp.PromptFeedback != nil && resp.PromptFeedback.BlockReason != "" {
			reason = "prompt blocked: " + resp.PromptFeedback.BlockReason
		}
		return nil, fmt.Errorf("gemini returned no content (%s)", reason)
	}

	candidate := resp.Candidates[0]
	var toolCalls []core.ToolCall
	var textContent strings.Builder
	var reasoning strings.Builder

	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			if part.Thought {
				reasoning.WriteString(part.Text)
			} else {
				textContent.WriteString(part.Text)
			}
		} else if part.FunctionCall != nil {
			args := "{}"
			if part.FunctionCall.Args != nil {
				raw, err := json.Marshal(part.FunctionCall.Args)
				if err != nil {
					continue
				}
				args = string(raw)
			}

			toolCall := core.ToolCall{
				ID:               part.FunctionCall.Name,
				Type:             "function",
				ThoughtSignature: part.ThoughtSignature,
			}
			toolCall.Function.Name = part.FunctionCall.Name
			toolCall.Function.Arguments = args
			toolCalls = append(toolCalls, toolCall)
		}
	}

	text := textContent.String()
	if text == "" && len(toolCalls) == 0 && candidate.FinishReason != "" && candidate.FinishReason != "STOP" {
		return nil, fmt.Errorf("gemini returned no content (finishReason: %s)", candidate.FinishReason)
	}

	output.Choices[0].Message = core.Message{
		Role:             "assistant",
		Content:          text,
		ReasoningContent: reasoning.String(),
		ToolCalls:        toolCalls,
	}
	output.Choices[0].FinishReason = candidate.FinishReason

	return output, nil
}
