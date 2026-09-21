package mistral

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	modelsAPI = "https://api.mistral.ai/v1/models"
)

func Models(ctx context.Context, config llmrouter.Config, filter llmrouter.ModelFilter) ([]string, error) {
	infos, err := ModelInfos(ctx, config, filter)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(infos))
	for _, m := range infos {
		ids = append(ids, m.ID)
	}
	return ids, nil
}

func ModelInfos(ctx context.Context, config llmrouter.Config, filter llmrouter.ModelFilter) ([]llmrouter.ModelInfo, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Models: APIKey is required")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	data, status, err := go_pkg_http.GET[llmrouter.MistralModels](ctx, client, modelsAPI, map[string]string{
		"Authorization": "Bearer " + config.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: http %d", status)
	}

	infos := make([]llmrouter.ModelInfo, 0, len(data.Data))
	for _, m := range data.Data {
		id := strings.TrimSpace(m.ID)
		if id == "" || !m.Capabilities.CompletionChat {
			continue
		}
		if filter.TextOnly && !llmrouter.IsTextModel(id) {
			continue
		}

		info := llmrouter.ModelInfo{ID: id, Thinking: m.Capabilities.Reasoning}
		if m.Capabilities.Reasoning {
			// * only none and high
			info.Efforts = []string{"none", "high"}
		}
		infos = append(infos, info)
	}
	return infos, nil
}
