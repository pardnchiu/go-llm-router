package gemini

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

const label = "gemini"

func (a *Agent) SendStream(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, _ llmrouter.Mode) (<-chan llmrouter.StreamEvent, error) {
	apiURL := fmt.Sprintf("%s%s:streamGenerateContent?alt=sse", baseAPI, a.model)

	resp, err := llmrouter.OpenStream(ctx, a.httpClient, apiURL, a.headers(), a.buildRequest(ctx, messages, tools, reasoning), label)
	if err != nil {
		return nil, err
	}

	events := make(chan llmrouter.StreamEvent)
	go func() {
		defer resp.Body.Close()
		defer close(events)

		emitted := false
		toolIndex := 0
		finishReason := ""
		var usage *llmrouter.Usage

		reader := bufio.NewReader(io.LimitReader(resp.Body, llmrouter.StreamBodyLimit))
		readErr := llmrouter.ScanSSE(reader, func(_, data string) bool {
			var chunk Output
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventError, Err: fmt.Errorf("%s stream decode: %w: %s", label, err, llmrouter.TruncateFrame(data))}
				return false
			}

			if chunk.PromptFeedback != nil && chunk.PromptFeedback.BlockReason != "" {
				events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventError, Err: fmt.Errorf("%s: prompt blocked: %s", label, chunk.PromptFeedback.BlockReason)}
				return false
			}

			if chunk.UsageMetadata != nil {
				usage = &llmrouter.Usage{
					Input:     chunk.UsageMetadata.PromptTokenCount - chunk.UsageMetadata.CachedContentTokenCount,
					Output:    chunk.UsageMetadata.CandidatesTokenCount,
					CacheRead: chunk.UsageMetadata.CachedContentTokenCount,
				}
			}

			for _, candidate := range chunk.Candidates {
				if candidate.FinishReason != "" {
					finishReason = candidate.FinishReason
				}
				for _, part := range candidate.Content.Parts {
					switch {
					case part.Text != "" && part.Thought:
						emitted = true
						events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventReasoning, ReasoningDelta: part.Text}

					case part.Text != "":
						emitted = true
						events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventText, TextDelta: part.Text}

					case part.FunctionCall != nil:
						emitted = true
						args := "{}"
						if part.FunctionCall.Args != nil {
							if raw, err := json.Marshal(part.FunctionCall.Args); err == nil {
								args = string(raw)
							}
						}
						events <- llmrouter.StreamEvent{
							Type: llmrouter.StreamEventToolCall,
							ToolCall: &llmrouter.ToolCallDelta{
								Index:            toolIndex,
								ID:               part.FunctionCall.Name,
								Name:             part.FunctionCall.Name,
								Arguments:        args,
								ThoughtSignature: part.ThoughtSignature,
							},
						}
						toolIndex++
					}
				}
			}
			return true
		})

		if readErr != nil {
			events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventError, Err: fmt.Errorf("%s stream read: %w", label, readErr)}
			return
		}

		if !emitted {
			switch {
			case finishReason != "" && finishReason != "STOP":
				events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventError, Err: fmt.Errorf("%s returned no content (finishReason: %s)", label, finishReason)}
			case finishReason == "":
				events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventError, Err: fmt.Errorf("%s returned no content (no candidates)", label)}
			}
			if finishReason != "STOP" {
				return
			}
		}

		if usage != nil {
			events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventUsage, Usage: usage}
		}
		events <- llmrouter.StreamEvent{Type: llmrouter.StreamEventDone, FinishReason: finishReason}
	}()

	return events, nil
}
