package ollamacloud

import (
	"context"
	"fmt"
	"net/http"
	"time"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const usageAPI = "https://ollama.com/api/usage"

type usageResponse struct {
	Limits struct {
		Monthly struct {
			Usage *float64 `json:"usage"`
		} `json:"monthly"`
	} `json:"limits"`
}

func Usage(ctx context.Context, config llmrouter.Config) (llmrouter.UsageRemaining, error) {
	if config.APIKey == "" {
		return llmrouter.UsageRemaining{}, fmt.Errorf("Usage: APIKey is required")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	data, status, err := go_pkg_http.GET[usageResponse](ctx, client, usageAPI, map[string]string{
		"Authorization": "Bearer " + config.APIKey,
	})
	if err != nil {
		return llmrouter.UsageRemaining{}, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: %w", err)
	}
	if status != http.StatusOK {
		return llmrouter.UsageRemaining{}, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: http %d", status)
	}
	if data.Limits.Monthly.Usage == nil {
		return llmrouter.UsageRemaining{}, fmt.Errorf("no limits.monthly.usage returned")
	}

	total := (1 - *data.Limits.Monthly.Usage) * 100
	return llmrouter.UsageRemaining{Total: &total}, nil
}
