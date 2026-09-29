package openaicodex

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	go_pkg_http "github.com/pardnchiu/go-pkg/http"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

const (
	modelsAPI = "https://models.agenvoy.com?target=codex"
)

func Models(ctx context.Context, _ llmrouter.Config, filter llmrouter.ModelFilter) ([]string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	data, status, err := go_pkg_http.GET[llmrouter.Models](ctx, client, modelsAPI, nil)
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
		if filter.TextOnly && !llmrouter.IsTextModel(id) {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}
