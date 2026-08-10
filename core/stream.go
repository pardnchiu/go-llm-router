package core

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const streamBodyLimit = 64 << 20

var ErrStreamUnsupported = errors.New("upstream does not support streaming")

func OpenStream(ctx context.Context, client *http.Client, url string, headers map[string]string, body map[string]any, label string) (*http.Response, int, error) {
	headers["Accept"] = "text/event-stream"
	headers["Accept-Encoding"] = "identity"

	resp, err := go_pkg_http.POSTStream(ctx, client, url, headers, body, "json")
	if err != nil {
		return nil, 0, fmt.Errorf("%s stream: %w", label, err)
	}
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
		return nil, resp.StatusCode, fmt.Errorf("%s stream: http %d: %s", label, resp.StatusCode, strings.TrimSpace(string(raw)))
	}
	if ct := resp.Header.Get("Content-Type"); ct != "" && !strings.Contains(ct, "text/event-stream") {
		defer resp.Body.Close()
		return nil, resp.StatusCode, fmt.Errorf("%s stream: got Content-Type %q: %w", label, ct, ErrStreamUnsupported)
	}
	return resp, resp.StatusCode, nil
}

type chatChunk struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
			Reasoning        string `json:"reasoning"`
			ToolCalls        []struct {
				Index    int    `json:"index"`
				ID       string `json:"id"`
				Function struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				} `json:"function"`
			} `json:"tool_calls"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage json.RawMessage `json:"usage"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func StreamChat(resp *http.Response, label string) <-chan StreamEvent {
	events := make(chan StreamEvent)

	go func() {
		defer resp.Body.Close()
		defer close(events)

		reader := bufio.NewReader(io.LimitReader(resp.Body, streamBodyLimit))
		readErr := ScanSSE(reader, func(_, data string) bool {
			if strings.TrimSpace(data) == "[DONE]" {
				return false
			}
			return handleChatChunk(data, label, events)
		})
		if readErr != nil {
			events <- StreamEvent{Type: StreamEventError, Err: fmt.Errorf("%s stream read: %w", label, readErr)}
		}
	}()

	return events
}

func handleChatChunk(data, label string, events chan<- StreamEvent) bool {
	var c chatChunk
	if err := json.Unmarshal([]byte(data), &c); err != nil {
		events <- StreamEvent{Type: StreamEventError, Err: fmt.Errorf("%s stream decode: %w", label, err)}
		return false
	}
	if c.Error != nil {
		events <- StreamEvent{Type: StreamEventError, Err: fmt.Errorf("%s", c.Error.Message)}
		return false
	}

	if len(c.Choices) > 0 {
		choice := c.Choices[0]
		if choice.Delta.Content != "" {
			events <- StreamEvent{Type: StreamEventText, TextDelta: choice.Delta.Content}
		}
		if reasoning := choice.Delta.ReasoningContent + choice.Delta.Reasoning; reasoning != "" {
			events <- StreamEvent{Type: StreamEventReasoning, ReasoningDelta: reasoning}
		}
		for _, tc := range choice.Delta.ToolCalls {
			events <- StreamEvent{
				Type: StreamEventToolCall,
				ToolCall: &ToolCallDelta{
					Index:     tc.Index,
					ID:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
		if choice.FinishReason != "" {
			events <- StreamEvent{Type: StreamEventDone, FinishReason: choice.FinishReason}
		}
	}

	if len(c.Usage) > 0 && string(c.Usage) != "null" {
		var usage Usage
		if err := json.Unmarshal(c.Usage, &usage); err == nil {
			events <- StreamEvent{Type: StreamEventUsage, Usage: &usage}
		}
	}

	return true
}

type responsesEvent struct {
	Type        string `json:"type"`
	Delta       string `json:"delta"`
	OutputIndex int    `json:"output_index"`
	Message     string `json:"message"`
	Item        *struct {
		Type      string `json:"type"`
		CallID    string `json:"call_id"`
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
		Summary   []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"summary"`
	} `json:"item"`
	Response *struct {
		Usage struct {
			InputTokens        int `json:"input_tokens"`
			OutputTokens       int `json:"output_tokens"`
			InputTokensDetails struct {
				CachedTokens int `json:"cached_tokens"`
			} `json:"input_tokens_details"`
		} `json:"usage"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	} `json:"response"`
}

type responsesState struct {
	sawToolCall    bool
	sawReasonDelta bool
}

func StreamResponses(resp *http.Response, label string) <-chan StreamEvent {
	events := make(chan StreamEvent)

	go func() {
		defer resp.Body.Close()
		defer close(events)

		state := responsesState{}
		reader := bufio.NewReader(io.LimitReader(resp.Body, streamBodyLimit))
		readErr := ScanSSE(reader, func(_, data string) bool {
			if strings.TrimSpace(data) == "[DONE]" {
				return false
			}
			return handleResponsesEvent(data, label, &state, events)
		})
		if readErr != nil {
			events <- StreamEvent{Type: StreamEventError, Err: fmt.Errorf("%s stream read: %w", label, readErr)}
		}
	}()

	return events
}

func handleResponsesEvent(data, label string, state *responsesState, events chan<- StreamEvent) bool {
	var ev responsesEvent
	if err := json.Unmarshal([]byte(data), &ev); err != nil {
		events <- StreamEvent{Type: StreamEventError, Err: fmt.Errorf("%s stream decode: %w", label, err)}
		return false
	}

	switch ev.Type {
	case "response.output_text.delta":
		events <- StreamEvent{Type: StreamEventText, TextDelta: ev.Delta}

	case "response.reasoning_summary_text.delta", "response.reasoning_text.delta":
		state.sawReasonDelta = true
		events <- StreamEvent{Type: StreamEventReasoning, ReasoningDelta: ev.Delta}

	case "response.output_item.added":
		if ev.Item != nil && ev.Item.Type == "function_call" {
			state.sawToolCall = true
			events <- StreamEvent{
				Type: StreamEventToolCall,
				ToolCall: &ToolCallDelta{
					Index:     ev.OutputIndex,
					ID:        ev.Item.CallID,
					Name:      ev.Item.Name,
					Arguments: ev.Item.Arguments,
				},
			}
		}

	case "response.function_call_arguments.delta":
		events <- StreamEvent{
			Type: StreamEventToolCall,
			ToolCall: &ToolCallDelta{
				Index:     ev.OutputIndex,
				Arguments: ev.Delta,
			},
		}

	case "response.output_item.done":
		if ev.Item == nil {
			break
		}
		if ev.Item.Type == "reasoning" && !state.sawReasonDelta {
			for _, s := range ev.Item.Summary {
				if s.Text != "" {
					events <- StreamEvent{Type: StreamEventReasoning, ReasoningDelta: s.Text}
				}
			}
		}
		if ev.Item.Type == "function_call" && !state.sawToolCall {
			state.sawToolCall = true
			events <- StreamEvent{
				Type: StreamEventToolCall,
				ToolCall: &ToolCallDelta{
					Index:     ev.OutputIndex,
					ID:        ev.Item.CallID,
					Name:      ev.Item.Name,
					Arguments: ev.Item.Arguments,
				},
			}
		}

	case "response.failed", "error":
		msg := ev.Message
		if ev.Response != nil && ev.Response.Error != nil {
			msg = ev.Response.Error.Message
		}
		if msg == "" {
			msg = ev.Type
		}
		events <- StreamEvent{Type: StreamEventError, Err: fmt.Errorf("%s stream: %s", label, msg)}
		return false

	case "response.completed", "response.incomplete":
		if ev.Response != nil {
			events <- StreamEvent{Type: StreamEventUsage, Usage: &Usage{
				Input:     ev.Response.Usage.InputTokens - ev.Response.Usage.InputTokensDetails.CachedTokens,
				Output:    ev.Response.Usage.OutputTokens,
				CacheRead: ev.Response.Usage.InputTokensDetails.CachedTokens,
			}}
		}
		finishReason := "stop"
		if state.sawToolCall {
			finishReason = "tool_calls"
		}
		events <- StreamEvent{Type: StreamEventDone, FinishReason: finishReason}
		return false
	}

	return true
}
