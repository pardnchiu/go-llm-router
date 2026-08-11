package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pardnchiu/go-llm-router/core"
)

const label = "claude"

func (a *Agent) SendStream(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning core.Reasoning, mode core.Mode) (<-chan core.StreamEvent, error) {
	requestBody := a.buildRequestBody(messages, tools, reasoning)
	requestBody["stream"] = true

	fast := a.applyMode(requestBody, mode)

	resp, err := core.OpenStream(ctx, a.httpClient, messagesAPI, a.headers(fast), requestBody, label)
	if err != nil {
		return nil, err
	}

	events := make(chan core.StreamEvent)
	go func() {
		defer resp.Body.Close()
		defer close(events)

		usage := core.Usage{}

		reader := bufio.NewReader(io.LimitReader(resp.Body, core.StreamBodyLimit))
		readErr := core.ScanSSE(reader, func(eventName, data string) bool {
			return a.handleClaudeSSE(eventName, data, &usage, events, fast)
		})
		if readErr != nil {
			events <- core.StreamEvent{Type: core.StreamEventError, Err: fmt.Errorf("%s stream read: %w", label, readErr)}
		}
	}()

	return events, nil
}

func (a *Agent) handleClaudeSSE(eventName, data string, usage *core.Usage, events chan<- core.StreamEvent, fast bool) bool {
	var evt streamEvent
	if err := json.Unmarshal([]byte(data), &evt); err != nil {
		events <- core.StreamEvent{Type: core.StreamEventError, Err: fmt.Errorf("%s stream decode: %w: %s", label, err, core.TruncateFrame(data))}
		return false
	}
	if eventName == "" {
		eventName = evt.Type
	}

	switch eventName {
	case "message_start":
		usage.Input = evt.Message.Usage.InputTokens
		usage.CacheCreate = evt.Message.Usage.CacheCreationInputTokens
		usage.CacheRead = evt.Message.Usage.CacheReadInputTokens
		if fast {
			core.WarnFastDowngrade("claude", a.model, evt.Message.Usage.Speed)
		}

	case "content_block_start":
		if evt.ContentBlock.Type == "tool_use" {
			events <- core.StreamEvent{
				Type: core.StreamEventToolCall,
				ToolCall: &core.ToolCallDelta{
					Index: evt.Index,
					ID:    evt.ContentBlock.ID,
					Name:  evt.ContentBlock.Name,
				},
			}
		}

	case "content_block_delta":
		switch evt.Delta.Type {
		case "text_delta":
			events <- core.StreamEvent{Type: core.StreamEventText, TextDelta: evt.Delta.Text}
		case "thinking_delta":
			events <- core.StreamEvent{Type: core.StreamEventReasoning, ReasoningDelta: evt.Delta.Thinking}
		case "input_json_delta":
			events <- core.StreamEvent{
				Type: core.StreamEventToolCall,
				ToolCall: &core.ToolCallDelta{
					Index:     evt.Index,
					Arguments: evt.Delta.PartialJSON,
				},
			}
		}

	case "message_delta":
		usage.Output = evt.Usage.OutputTokens
		events <- core.StreamEvent{Type: core.StreamEventUsage, Usage: usage}
		if evt.Delta.StopReason != "" {
			events <- core.StreamEvent{Type: core.StreamEventDone, FinishReason: evt.Delta.StopReason}
		}

	case "error":
		msg := evt.Error.Message
		if msg == "" {
			msg = evt.Error.Type
		}
		if msg == "" {
			msg = core.TruncateFrame(data)
		}
		events <- core.StreamEvent{Type: core.StreamEventError, Err: fmt.Errorf("%s stream: %s", label, msg)}
		return false

	case "message_stop":
		return false
	}

	return true
}
