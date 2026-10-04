package openrouter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	go_pkg_http "github.com/pardnchiu/go-pkg/http"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

const creditsAPI = "https://openrouter.ai/api/v1/credits"

type creditsResponse struct {
	Data struct {
		TotalCredits float64 `json:"total_credits"`
		TotalUsage   float64 `json:"total_usage"`
	} `json:"data"`
}

func Usage(ctx context.Context, config llmrouter.Config) (llmrouter.UsageRemaining, error) {
	if config.APIKey == "" {
		return llmrouter.UsageRemaining{}, fmt.Errorf("Usage: APIKey is required")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	data, status, err := go_pkg_http.GET[creditsResponse](ctx, client, creditsAPI, map[string]string{
		"Authorization": "Bearer " + config.APIKey,
	})
	if err != nil {
		return llmrouter.UsageRemaining{}, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: %w", err)
	}
	if status != http.StatusOK {
		return llmrouter.UsageRemaining{}, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: http %d", status)
	}

	balance := data.Data.TotalCredits - data.Data.TotalUsage
	return llmrouter.UsageRemaining{Balance: &balance}, nil
}
