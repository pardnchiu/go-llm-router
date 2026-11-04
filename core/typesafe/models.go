package typesafe

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	go_pkg_http "github.com/pardnchiu/go-pkg/http"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

const modelsAPI = "https://api.typesafe.ai/v1/models"

func Models(ctx context.Context, config llmrouter.Config) ([]string, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("[Denied] typesafe.Models: APIKey is required")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	data, status, err := go_pkg_http.GET[struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}](ctx, client, modelsAPI, map[string]string{
		"Authorization": "Bearer " + config.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("[%s] typesafe.Models: %w", llmrouter.HTTPErrorLabel(status, err), err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("[%s] typesafe.Models: http %d", llmrouter.HTTPErrorLabel(status, nil), status)
	}

	names := make([]string, 0, len(data.Models))
	for _, m := range data.Models {
		if name := strings.TrimSpace(m.Name); name != "" {
			names = append(names, name)
		}
	}
	return names, nil
}
