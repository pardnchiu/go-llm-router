package openaicodex

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	copilotResponse "github.com/pardnchiu/go-llm-router/core/copilot/response"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	responsesAPI      = "https://chatgpt.com/backend-api/codex/responses"
	promptCachePrefix = "agenvoy-"
	promptCacheKeyLen = 24
)

func (a *Agent) headers(ctx context.Context) (map[string]string, error) {
	auth, err := a.authHeader(ctx)
	if err != nil {
		return nil, fmt.Errorf("a.authHeader: %w", err)
	}

	headers := map[string]string{
		"Authorization": auth,
		"Content-Type":  "application/json",
	}
	if a.token != nil && a.token.AccountID != "" {
		headers["ChatGPT-Account-Id"] = a.token.AccountID
	}
	return headers, nil
}

func (a *Agent) buildBody(messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning) map[string]any {
	var instructions string
	var nonSystem []llmrouter.Message
	for _, m := range messages {
		if m.Role == "system" {
			if s, ok := m.Content.(string); ok {
				if instructions != "" {
					instructions += "\n"
				}
				instructions += s
			}
		} else {
			nonSystem = append(nonSystem, m)
		}
	}

	body := map[string]any{
		"model":        a.model,
		"input":        copilotResponse.ConvertInput(nonSystem),
		"tools":        copilotResponse.ConvertTools(tools),
		"instructions": instructions,
		"store":        false,
		"stream":       true,
	}
	if effort, ok := a.effort(reasoning); ok {
		body["reasoning"] = map[string]any{"effort": effort, "summary": "auto"}
	}
	if key := promptCacheKey(instructions); key != "" {
		body["prompt_cache_key"] = key
	}
	return body
}

func (a *Agent) Send(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (*llmrouter.Output, int, error) {
	headers, err := a.headers(ctx)
	if err != nil {
		return nil, 0, err
	}

	resp, err := go_pkg_http.POSTStream(ctx, a.httpClient, responsesAPI, headers, a.buildBody(messages, tools, reasoning), "json")
	if err != nil {
		return nil, 0, fmt.Errorf("github.com/pardnchiu/go-pkg/http: POSTStream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, llmrouter.ErrorBodyLimit))
		return nil, resp.StatusCode, &llmrouter.StreamError{
			Provider: label,
			Code:     resp.StatusCode,
			Body:     strings.TrimSpace(string(raw)),
		}
	}

	out, err := parseSSEStream(resp)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return out, resp.StatusCode, nil
}

func promptCacheKey(instructions string) string {
	if instructions == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(instructions))
	return promptCachePrefix + hex.EncodeToString(sum[:])[:promptCacheKeyLen]
}

type sseEvent struct {
	Type      string `json:"type"`
	Delta     string `json:"delta"`
	ItemID    string `json:"item_id"`
	Arguments string `json:"arguments"`
	Item      *struct {
		ID        string `json:"id"`
		Type      string `json:"type"`
		CallID    string `json:"call_id"`
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
		Summary   []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"summary"`
	} `json:"item"`
	Response *copilotResponse.Output `json:"response"`
}

type pendingCall struct {
	itemID string
	callID string
	name   string
	args   string
}

func parseSSEStream(resp *http.Response) (*llmrouter.Output, error) {
	var (
		textBuf        strings.Builder
		completedText  string
		reasonDeltaBuf strings.Builder
		reasonItemBuf  strings.Builder
		toolCalls      []llmrouter.ToolCall
		usage          llmrouter.Usage
		argsBuf        = map[string]*strings.Builder{}
		pending        []pendingCall
	)
	getBuf := func(key string) *strings.Builder {
		if key == "" {
			return nil
		}
		b, ok := argsBuf[key]
		if !ok {
			b = &strings.Builder{}
			argsBuf[key] = b
		}
		return b
	}

	var streamErr error
	handle := func(_, data string) bool {
		if strings.TrimSpace(data) == "[DONE]" {
			return false
		}

		var ev sseEvent
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			slog.Debug("dropped undecodable SSE frame",
				slog.String("provider", label), slog.String("err", err.Error()))
			return true
		}
		if ev.Type == "response.output_item.done" && ev.Item != nil && ev.Item.Type == "reasoning" {
			for _, s := range ev.Item.Summary {
				reasonItemBuf.WriteString(s.Text)
			}
		}

		switch ev.Type {
		case "response.output_text.delta":
			textBuf.WriteString(ev.Delta)

		case "response.reasoning_summary_text.delta", "response.reasoning_text.delta":
			reasonDeltaBuf.WriteString(ev.Delta)

		case "response.function_call_arguments.delta":
			if b := getBuf(ev.ItemID); b != nil {
				b.WriteString(ev.Delta)
			}

		case "response.function_call_arguments.done":
			if ev.Arguments != "" {
				if b := getBuf(ev.ItemID); b != nil {
					b.Reset()
					b.WriteString(ev.Arguments)
				}
			}

		case "response.output_item.added":
			if ev.Item != nil && ev.Item.Type == "function_call" {
				pending = append(pending, pendingCall{
					itemID: ev.Item.ID,
					callID: ev.Item.CallID,
					name:   ev.Item.Name,
					args:   ev.Item.Arguments,
				})
			}

		case "response.output_item.done":
			if ev.Item != nil && ev.Item.Type == "function_call" {
				found := false
				for i := range pending {
					if pending[i].itemID == ev.Item.ID || (pending[i].callID != "" && pending[i].callID == ev.Item.CallID) {
						if ev.Item.Arguments != "" {
							pending[i].args = ev.Item.Arguments
						}
						if pending[i].name == "" {
							pending[i].name = ev.Item.Name
						}
						if pending[i].callID == "" {
							pending[i].callID = ev.Item.CallID
						}
						found = true
						break
					}
				}
				if !found {
					pending = append(pending, pendingCall{
						itemID: ev.Item.ID,
						callID: ev.Item.CallID,
						name:   ev.Item.Name,
						args:   ev.Item.Arguments,
					})
				}
			}

		case "response.failed", "error":
			streamErr = llmrouter.ResponsesStreamError(label, data)
			return false

		case "response.completed", "response.incomplete":
			if ev.Response != nil {
				usage = llmrouter.Usage{
					Input:     ev.Response.Usage.InputTokens - ev.Response.Usage.InputTokensDetails.CachedTokens,
					Output:    ev.Response.Usage.OutputTokens,
					CacheRead: ev.Response.Usage.InputTokensDetails.CachedTokens,
				}
				out := copilotResponse.ConvertOutput(*ev.Response)
				if len(out.Choices) > 0 {
					if str, ok := out.Choices[0].Message.Content.(string); ok {
						completedText = str
					}
					if len(pending) == 0 {
						toolCalls = out.Choices[0].Message.ToolCalls
					}
				}
			}
		}
		return true
	}

	if err := llmrouter.ScanSSE(bufio.NewReader(io.LimitReader(resp.Body, llmrouter.StreamBodyLimit)), handle); err != nil {
		return nil, fmt.Errorf("codex stream read: %w", err)
	}
	if streamErr != nil {
		return nil, streamErr
	}

	for _, p := range pending {
		args := p.args
		if args == "" {
			if b, ok := argsBuf[p.itemID]; ok {
				args = b.String()
			}
		}
		if args == "" && len(argsBuf) == 1 {
			for _, b := range argsBuf {
				args = b.String()
			}
		}
		toolCalls = append(toolCalls, llmrouter.ToolCall{
			ID:   p.callID,
			Type: "function",
			Function: struct {
				Name      string `json:"name"`
				Arguments string `json:"arguments"`
			}{
				Name:      p.name,
				Arguments: args,
			},
		})
	}

	msg := llmrouter.Message{Role: "assistant"}
	if str := textBuf.String(); str != "" {
		msg.Content = str
	} else if completedText != "" {
		msg.Content = completedText
	}
	msg.ReasoningContent = reasonDeltaBuf.String()
	if reasonItemBuf.Len() > len(msg.ReasoningContent) {
		msg.ReasoningContent = reasonItemBuf.String()
	}
	msg.ToolCalls = toolCalls

	finishReason := "stop"
	if len(toolCalls) > 0 {
		finishReason = "tool_calls"
	}

	return &llmrouter.Output{
		Choices: []llmrouter.OutputChoices{
			{Message: msg, FinishReason: finishReason},
		},
		Usage: usage,
	}, nil
}
