package ollamacloud

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pardnchiu/go-llm-router/core"
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

func Usage(ctx context.Context, config core.Config) (float64, error) {
	if config.APIKey == "" {
		return 0, fmt.Errorf("Usage: APIKey is required")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	data, status, err := go_pkg_http.GET[usageResponse](ctx, client, usageAPI, map[string]string{
		"Authorization": "Bearer " + config.APIKey,
	})
	if err != nil {
		return 0, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: %w", err)
	}
	if status != http.StatusOK {
		return 0, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: http %d", status)
	}
	if data.Limits.Monthly.Usage == nil {
		return 0, fmt.Errorf("no limits.monthly.usage returned")
	}

	return (1 - *data.Limits.Monthly.Usage) * 100, nil
}
