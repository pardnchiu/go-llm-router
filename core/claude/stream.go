package claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

func (a *Agent) SendStream(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning core.Reasoning, mode core.Mode) (<-chan core.StreamEvent, error) {
	requestBody := a.buildRequestBody(messages, tools, reasoning)
	requestBody["stream"] = true

	fast := a.applyMode(requestBody, mode)
	headers := a.headers(fast)
	headers["Accept"] = "text/event-stream"
	headers["Accept-Encoding"] = "identity"

	resp, err := go_pkg_http.POSTStream(ctx, a.httpClient, messagesAPI, headers, requestBody, "json")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return nil, fmt.Errorf("claude stream: http %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	events := make(chan core.StreamEvent)
	go func() {
		defer resp.Body.Close()
		defer close(events)

		usage := core.Usage{}

		reader := bufio.NewReader(io.LimitReader(resp.Body, 64<<20))
		var eventName string
		for {
			line, readErr := reader.ReadString('\n')
			line = strings.TrimRight(line, "\r\n")

			switch {
			case strings.HasPrefix(line, "event:"):
				eventName = strings.TrimSpace(strings.TrimPrefix(line, "event:"))
			case strings.HasPrefix(line, "data:"):
				data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
				if data != "" {
					if !a.handleClaudeSSE(eventName, data, &usage, events, fast) {
						return
					}
				}
			}

			if readErr != nil {
				if readErr != io.EOF {
					events <- core.StreamEvent{Type: core.StreamEventError, Err: fmt.Errorf("claude stream read: %w", readErr)}
				}
				return
			}
		}
	}()

	return events, nil
}

func (a *Agent) handleClaudeSSE(eventName, data string, usage *core.Usage, events chan<- core.StreamEvent, fast bool) bool {
	var evt streamEvent
	if err := json.Unmarshal([]byte(data), &evt); err != nil {
		events <- core.StreamEvent{Type: core.StreamEventError, Err: fmt.Errorf("claude stream decode: %w", err)}
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
		events <- core.StreamEvent{Type: core.StreamEventError, Err: fmt.Errorf("%s", evt.Error.Message)}
		return false

	case "message_stop":
		return false
	}

	return true
}
