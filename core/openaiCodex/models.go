package openaicodex

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	go_pkg_http "github.com/pardnchiu/go-pkg/http"

	"github.com/pardnchiu/go-llm-router/core"
)

const (
	modelsAPI = "https://agenvoy-codex.pardn.workers.dev/models"
)

func Models(ctx context.Context, config core.Config, filter core.ModelFilter) ([]string, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Models: APIKey is required")
	}

	headers := map[string]string{
		"Authorization": "Bearer " + config.APIKey,
	}
	if config.AccountID != "" {
		headers["ChatGPT-Account-Id"] = config.AccountID
	}

	client := &http.Client{Timeout: 10 * time.Second}
	data, status, err := go_pkg_http.GET[core.Models](ctx, client, modelsAPI, headers)
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: http %d", status)
	}

	ids := make([]string, 0, len(data.Data))
	for _, m := range data.Data {
		id := strings.TrimSpace(m.ID)
		if id == "" {
			continue
		}
		if filter.TextOnly && !core.IsTextModel(id) {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}
