package mistral

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
)

const label = "mistral"

type streamChunk struct {
	Choices []struct {
		Delta struct {
			Content   json.RawMessage `json:"content"`
			ToolCalls []struct {
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
}

func (a *Agent) SendStream(ctx context.Context, messages []core.Message, tools []core.Tool, reasoning core.Reasoning, mode core.Mode) (<-chan core.StreamEvent, error) {
	body := a.buildBody(messages, tools, reasoning)
	body["stream"] = true
	resp, err := core.OpenStream(ctx, a.httpClient, chatAPI, a.headers(), body, label)
	if err != nil {
		return nil, err
	}
	return streamEvents(resp), nil
}

func streamEvents(resp *http.Response) <-chan core.StreamEvent {
	events := make(chan core.StreamEvent)

	go func() {
		defer resp.Body.Close()
		defer close(events)

		reader := bufio.NewReader(io.LimitReader(resp.Body, core.StreamBodyLimit))
		readErr := core.ScanSSE(reader, func(_, data string) bool {
			if strings.TrimSpace(data) == "[DONE]" {
				return false
			}
			return handleChunk(data, events)
		})
		if readErr != nil {
			events <- core.StreamEvent{Type: core.StreamEventError, Err: fmt.Errorf("%s stream read: %w", label, readErr)}
		}
	}()

	return events
}

func handleChunk(data string, events chan<- core.StreamEvent) bool {
	var c streamChunk
	if err := json.Unmarshal([]byte(data), &c); err != nil {
		events <- core.StreamEvent{Type: core.StreamEventError, Err: fmt.Errorf("%s stream decode: %w: %s", label, err, core.TruncateFrame(data))}
		return false
	}

	if len(c.Choices) > 0 {
		choice := c.Choices[0]
		text, reasoning := splitContent(choice.Delta.Content)
		if text != "" {
			events <- core.StreamEvent{Type: core.StreamEventText, TextDelta: text}
		}
		if reasoning != "" {
			events <- core.StreamEvent{Type: core.StreamEventReasoning, ReasoningDelta: reasoning}
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

	if len(c.Usage) > 0 && string(c.Usage) != "null" {
		var usage core.Usage
		if err := json.Unmarshal(c.Usage, &usage); err == nil {
			events <- core.StreamEvent{Type: core.StreamEventUsage, Usage: &usage}
		}
	}

	return true
}
