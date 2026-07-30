package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

func Models(ctx context.Context, config core.Config, filter core.ModelFilter) ([]string, error) {
	if config.APIKey == "" || config.AccountID == "" {
		return nil, fmt.Errorf("Models: APIKey and AccountID are required")
	}

	endpoint := runAPI + config.AccountID + "/ai/models/search"
	client := &http.Client{Timeout: 10 * time.Second}
	data, status, err := go_pkg_http.GET[core.CloudFlareModels](ctx, client, endpoint, map[string]string{
		"Authorization": "Bearer " + config.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: http %d", status)
	}

	ids := make([]string, 0, len(data.Result))
	for _, m := range data.Result {
		if m.Name == "" || m.Task.Name != "Text Generation" {
			continue
		}
		if filter.TextOnly && !core.IsTextModel(m.Name) {
			continue
		}
		ids = append(ids, m.Name)
	}
	return ids, nil
}
