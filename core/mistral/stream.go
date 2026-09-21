package mistral

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
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

func (a *Agent) SendStream(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (<-chan llmrouter.StreamEvent, error) {
	body := a.buildBody(messages, tools, reasoning)
	body["stream"] = true
	resp, err := llmrouter.OpenStream(ctx, a.httpClient, chatAPI, a.headers(), body, label)
	if err != nil {
		return nil, err
	}
	return streamEvents(resp), nil
}

func streamEvents(resp *http.Response) <-chan llmrouter.StreamEvent {
	events := make(chan llmrouter.StreamEvent)

	go func() {
		defer resp.Body.Close()
		defer close(events)

		reader := bufio.NewReader(io.LimitReader(resp.Body, llmrouter.StreamBodyLimit))
		readErr := llmrouter.ScanSSE(reader, func(_, data string) bool {
			if strings.TrimSpace(data) == "[DONE]" {
				return false
			}
			return handleChunk(data, events)
		})
		if readErr != nil {
			events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventError, Err: fmt.Errorf("%s stream read: %w", label, readErr)}
		}
	}()

	return events
}

func handleChunk(data string, events chan<- llmrouter.StreamEvent) bool {
	var c streamChunk
	if err := json.Unmarshal([]byte(data), &c); err != nil {
		events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventError, Err: fmt.Errorf("%s stream decode: %w: %s", label, err, llmrouter.TruncateFrame(data))}
		return false
	}

	if len(c.Choices) > 0 {
		choice := c.Choices[0]
		text, reasoning := splitContent(choice.Delta.Content)
		if text != "" {
			events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventText, TextDelta: text}
		}
		if reasoning != "" {
			events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventReasoning, ReasoningDelta: reasoning}
		}
		for _, tc := range choice.Delta.ToolCalls {
			events <- llmrouter.StreamEvent{
				Type: llmrouter.StreamEventToolCall,
				ToolCall: &llmrouter.ToolCallDelta{
					Index:     tc.Index,
					ID:        tc.ID,
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
		if choice.FinishReason != "" {
			events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventDone, FinishReason: choice.FinishReason}
		}
	}

	if len(c.Usage) > 0 && string(c.Usage) != "null" {
		var usage llmrouter.Usage
		if err := json.Unmarshal(c.Usage, &usage); err == nil {
			events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventUsage, Usage: &usage}
		}
	}

	return true
}
