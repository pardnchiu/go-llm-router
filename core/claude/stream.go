package claude

import (
	"bufio"
	"context"
	"fmt"
	"io"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

const label = "claude"

func (a *Agent) SendStream(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (<-chan llmrouter.StreamEvent, error) {
	requestBody := a.buildRequestBody(messages, tools, reasoning)
	requestBody["stream"] = true
	for _, tool := range requestBody["tools"].([]map[string]any) {
		tool["eager_input_streaming"] = true
	}

	fast := a.applyMode(requestBody, mode)

	resp, err := llmrouter.OpenStream(ctx, a.httpClient, messagesAPI, a.headers(fast), requestBody, label)
	if err != nil {
		return nil, err
	}

	events := make(chan llmrouter.StreamEvent)
	go func() {
		defer resp.Body.Close()
		defer close(events)

		usage := llmrouter.Usage{}

		reader := bufio.NewReader(resp.Body)
		readErr := llmrouter.ScanSSE(reader, func(eventName, data string) bool {
			return a.handleClaudeSSE(eventName, data, &usage, events, fast, resp.Body)
		})
		if readErr != nil {
			events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventError, Err: fmt.Errorf("%s stream read: %w", label, readErr)}
		}
	}()

	return events, nil
}

func (a *Agent) handleClaudeSSE(eventName, data string, usage *llmrouter.Usage, events chan<- llmrouter.StreamEvent, fast bool, body io.Reader) bool {
	var evt streamEvent
	ok, err := llmrouter.DecodeFrame(body, label, data, &evt)
	if err != nil {
		events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventError, Err: err}
		return false
	}
	if !ok {
		return true
	}
	if eventName == "" {
		eventName = evt.Type
	}
	if eventName == "content_block_start" || eventName == "content_block_delta" {
		llmrouter.TouchIdle(body, evt.Delta.Text, evt.Delta.Thinking, evt.Delta.PartialJSON)
	}

	switch eventName {
	case "message_start":
		usage.Input = evt.Message.Usage.InputTokens
		usage.CacheCreate = evt.Message.Usage.CacheCreationInputTokens
		usage.CacheRead = evt.Message.Usage.CacheReadInputTokens
		if fast {
			llmrouter.WarnFastDowngrade("claude", a.model, evt.Message.Usage.Speed)
		}

	case "content_block_start":
		if evt.ContentBlock.Type == "tool_use" {
			events <- llmrouter.StreamEvent{
				Type: llmrouter.StreamEventToolCall,
				ToolCall: &llmrouter.ToolCallDelta{
					Index: evt.Index,
					ID:    evt.ContentBlock.ID,
					Name:  evt.ContentBlock.Name,
				},
			}
		}

	case "content_block_delta":
		switch evt.Delta.Type {
		case "text_delta":
			events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventText, TextDelta: evt.Delta.Text}
		case "thinking_delta":
			events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventReasoning, ReasoningDelta: evt.Delta.Thinking}
		case "input_json_delta":
			events <- llmrouter.StreamEvent{
				Type: llmrouter.StreamEventToolCall,
				ToolCall: &llmrouter.ToolCallDelta{
					Index:     evt.Index,
					Arguments: evt.Delta.PartialJSON,
				},
			}
		}

	case "message_delta":
		usage.Output = evt.Usage.OutputTokens
		events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventUsage, Usage: usage}
		if evt.Delta.StopReason != "" {
			events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventDone, FinishReason: evt.Delta.StopReason}
		}

	case "error":
		msg := evt.Error.Message
		if msg == "" {
			msg = evt.Error.Type
		}
		if msg == "" {
			msg = llmrouter.TruncateFrame(data)
		}
		events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventError, Err: fmt.Errorf("%s stream: %s", label, msg)}
		return false

	case "message_stop":
		return false
	}

	return true
}
