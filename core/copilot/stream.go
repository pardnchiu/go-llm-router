package copilot

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
	headers, err := a.headers(ctx)
	if err != nil {
		return nil, err
	}
	headers["Accept"] = "text/event-stream"
	headers["Accept-Encoding"] = "identity"

	if a.useResponses(ctx) {
		body := a.buildResponsesBody(messages, tools, reasoning)
		body["stream"] = true
		return a.streamResponses(ctx, headers, body)
	}

	body := a.buildChatBody(messages, tools, reasoning)
	body["stream"] = true
	body["stream_options"] = map[string]any{"include_usage": true}
	return a.streamChat(ctx, headers, body)
}

func (a *Agent) streamChat(ctx context.Context, headers map[string]string, body map[string]any) (<-chan core.StreamEvent, error) {
	resp, err := go_pkg_http.POSTStream(ctx, a.httpClient, chatAPI, headers, body, "json")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return nil, fmt.Errorf("copilot chat stream: http %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	events := make(chan core.StreamEvent)
	go func() {
		defer resp.Body.Close()
		defer close(events)

		reader := bufio.NewReader(io.LimitReader(resp.Body, 64<<20))
		for {
			line, readErr := reader.ReadString('\n')
			line = strings.TrimRight(line, "\r\n")

			if data, ok := strings.CutPrefix(line, "data:"); ok {
				data = strings.TrimSpace(data)
				if data != "" {
					if data == "[DONE]" {
						return
					}
					if !handleChatChunk(data, events) {
						return
					}
				}
			}

			if readErr != nil {
				if readErr != io.EOF {
					events <- core.StreamEvent{Type: core.StreamEventError, Err: fmt.Errorf("copilot chat stream read: %w", readErr)}
				}
				return
			}
		}
	}()

	return events, nil
}

func handleChatChunk(data string, events chan<- core.StreamEvent) bool {
	var c chatChunk
	if err := json.Unmarshal([]byte(data), &c); err != nil {
		events <- core.StreamEvent{Type: core.StreamEventError, Err: fmt.Errorf("copilot chat stream decode: %w", err)}
		return false
	}
	if c.Error != nil {
		events <- core.StreamEvent{Type: core.StreamEventError, Err: fmt.Errorf("%s", c.Error.Message)}
		return false
	}

	if len(c.Choices) > 0 {
		choice := c.Choices[0]
		if choice.Delta.Content != "" {
			events <- core.StreamEvent{Type: core.StreamEventText, TextDelta: choice.Delta.Content}
		}
		if choice.Delta.ReasoningContent != "" {
			events <- core.StreamEvent{Type: core.StreamEventReasoning, ReasoningDelta: choice.Delta.ReasoningContent}
		}
		for _, tc := range choice.Delta.ToolCalls {
			events <- core.StreamEvent{
				Type: core.StreamEventToolCall,
				ToolCall: &core.ToolCallDelta{
					Index:     tc.Index,
					ID:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
		if choice.FinishReason != "" {
			events <- core.StreamEvent{Type: core.StreamEventDone, FinishReason: choice.FinishReason}
		}
	}

	if len(c.Usage) > 0 {
		var usage core.Usage
		if err := json.Unmarshal(c.Usage, &usage); err == nil {
			events <- core.StreamEvent{Type: core.StreamEventUsage, Usage: &usage}
		}
	}

	return true
}

func (a *Agent) streamResponses(ctx context.Context, headers map[string]string, body map[string]any) (<-chan core.StreamEvent, error) {
	resp, err := go_pkg_http.POSTStream(ctx, a.httpClient, responsesAPI, headers, body, "json")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return nil, fmt.Errorf("copilot responses stream: http %d: %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	events := make(chan core.StreamEvent)
	go func() {
		defer resp.Body.Close()
		defer close(events)

		sawToolCall := false
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
					if !handleResponsesEvent(eventName, data, &sawToolCall, events) {
						return
					}
				}
			}

			if readErr != nil {
				if readErr != io.EOF {
					events <- core.StreamEvent{Type: core.StreamEventError, Err: fmt.Errorf("copilot responses stream read: %w", readErr)}
				}
				return
			}
		}
	}()

	return events, nil
}

func handleResponsesEvent(eventName, data string, sawToolCall *bool, events chan<- core.StreamEvent) bool {
	var evt responsesEvent
	if err := json.Unmarshal([]byte(data), &evt); err != nil {
		events <- core.StreamEvent{Type: core.StreamEventError, Err: fmt.Errorf("copilot responses stream decode: %w", err)}
		return false
	}
	if eventName == "" {
		eventName = evt.Type
	}

	switch eventName {
	case "response.output_text.delta":
		events <- core.StreamEvent{Type: core.StreamEventText, TextDelta: evt.Delta}

	case "response.reasoning_summary_text.delta":
		events <- core.StreamEvent{Type: core.StreamEventReasoning, ReasoningDelta: evt.Delta}

	case "response.output_item.added":
		if evt.Item.Type == "function_call" {
			*sawToolCall = true
			events <- core.StreamEvent{
				Type: core.StreamEventToolCall,
				ToolCall: &core.ToolCallDelta{
					Index: evt.OutputIndex,
					ID:    evt.Item.CallID,
					Name:  evt.Item.Name,
				},
			}
		}

	case "response.function_call_arguments.delta":
		events <- core.StreamEvent{
			Type: core.StreamEventToolCall,
			ToolCall: &core.ToolCallDelta{
				Index:     evt.OutputIndex,
				Arguments: evt.Delta,
			},
		}

	case "response.completed":
		usage := core.Usage{
			Input:     evt.Response.Usage.InputTokens - evt.Response.Usage.InputTokensDetails.CachedTokens,
			Output:    evt.Response.Usage.OutputTokens,
			CacheRead: evt.Response.Usage.InputTokensDetails.CachedTokens,
		}
		events <- core.StreamEvent{Type: core.StreamEventUsage, Usage: &usage}
		finishReason := "stop"
		if *sawToolCall {
			finishReason = "tool_calls"
		}
		events <- core.StreamEvent{Type: core.StreamEventDone, FinishReason: finishReason}
		return false

	case "response.failed", "error":
		msg := evt.Message
		if evt.Response.Error != nil {
			msg = evt.Response.Error.Message
		}
		events <- core.StreamEvent{Type: core.StreamEventError, Err: fmt.Errorf("%s", msg)}
		return false
	}

	return true
}
