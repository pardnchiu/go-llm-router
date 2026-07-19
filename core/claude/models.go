package claude

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	modelsAPI = "https://api.anthropic.com/v1/models"
)

func Models(ctx context.Context, config core.Config) ([]string, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Models: APIKey is required")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	data, status, err := go_pkg_http.GET[core.Models](ctx, client, modelsAPI, map[string]string{
		"x-api-key":         config.APIKey,
		"anthropic-version": "2023-06-01",
	})
	if err != nil {
		return nil, fmt.Errorf("go_pkg_http.GET: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("go_pkg_http.GET: http %d", status)
	}

	ids := make([]string, 0, len(data.Data))
	for _, m := range data.Data {
		if id := strings.TrimSpace(m.ID); id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
