package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"time"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

func Models(ctx context.Context, config llmrouter.Config, filter llmrouter.ModelFilter) ([]string, error) {
	if config.APIKey == "" || config.AccountID == "" {
		return nil, fmt.Errorf("Models: APIKey and AccountID are required")
	}

	endpoint := runAPI + config.AccountID + "/ai/models/search"
	client := &http.Client{Timeout: 10 * time.Second}
	data, status, err := go_pkg_http.GET[llmrouter.CloudFlareModels](ctx, client, endpoint, map[string]string{
		"Authorization": "Bearer " + config.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: http %d", status)
	}

	task := "Text Generation"
	switch {
	case filter.STTOnly:
		task = "Automatic Speech Recognition"
	case filter.TTSOnly:
		task = "Text-to-Speech"
	}

	ids := make([]string, 0, len(data.Result))
	for _, m := range data.Result {
		if m.Name == "" || m.Task.Name != task {
			continue
		}
		if filter.TextOnly && !llmrouter.IsTextModel(m.Name) {
			continue
		}
		ids = append(ids, m.Name)
	}
	return ids, nil
}
