package grok

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/xai"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

func (a *Agent) headers() map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + a.apiKey,
		"Content-Type":  "application/json",
	}
}

func (a *Agent) buildBody(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, fast bool) map[string]any {
	effort, _ := a.effort(reasoning)
	return xai.BuildBody(ctx, a.model, messages, tools, effort, fast)
}

func (a *Agent) Send(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (*llmrouter.Output, int, error) {
	fast := mode == llmrouter.ModeFast && llmrouter.SupportFast(label, a.model)

	resp, err := go_pkg_http.POSTStream(ctx, a.httpClient, xai.ResponsesAPI, a.headers(), a.buildBody(ctx, messages, tools, reasoning, fast), "json")
	if err != nil {
		return nil, 0, fmt.Errorf("go_pkg_http.POSTStream: %w", err)
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

	out, err := xai.ParseSSEStream(label, resp)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if fast {
		llmrouter.WarnFastDowngrade(label, a.model, out.ServiceTier)
	}
	return out, resp.StatusCode, nil
}
