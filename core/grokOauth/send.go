package grokoauth

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

func (a *Agent) headers(ctx context.Context) (map[string]string, error) {
	auth, err := a.authHeader(ctx)
	if err != nil {
		return nil, fmt.Errorf("a.authHeader: %w", err)
	}
	return map[string]string{
		"Authorization": auth,
		"Content-Type":  "application/json",
	}, nil
}

func (a *Agent) buildBody(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, fast bool) map[string]any {
	effort, _ := a.effort(reasoning)
	return xai.BuildBody(ctx, a.model, messages, tools, effort, fast)
}

func (a *Agent) Send(ctx context.Context, messages []llmrouter.Message, tools []llmrouter.Tool, reasoning llmrouter.Reasoning, mode llmrouter.Mode) (*llmrouter.Output, int, error) {
	headers, err := a.headers(ctx)
	if err != nil {
		return nil, 0, err
	}
	fast := mode == llmrouter.ModeFast && llmrouter.SupportFast("grok", a.model)

	resp, err := go_pkg_http.POSTStream(ctx, a.httpClient, xai.ResponsesAPI, headers, a.buildBody(ctx, messages, tools, reasoning, fast), "json")
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
		llmrouter.WarnFastDowngrade("grok-oauth", a.model, out.ServiceTier)
	}
	return out, resp.StatusCode, nil
}
