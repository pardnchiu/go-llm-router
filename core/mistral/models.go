package mistral

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
	modelsAPI = "https://api.mistral.ai/v1/models"
)

func Models(ctx context.Context, config core.Config, filter core.ModelFilter) ([]string, error) {
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

func ModelInfos(ctx context.Context, config core.Config, filter core.ModelFilter) ([]core.ModelInfo, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Models: APIKey is required")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	data, status, err := go_pkg_http.GET[core.MistralModels](ctx, client, modelsAPI, map[string]string{
		"Authorization": "Bearer " + config.APIKey,
	})
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: http %d", status)
	}

	infos := make([]core.ModelInfo, 0, len(data.Data))
	for _, m := range data.Data {
		id := strings.TrimSpace(m.ID)
		if id == "" || !m.Capabilities.CompletionChat {
			continue
		}
		if filter.TextOnly && !core.IsTextModel(id) {
			continue
		}

		info := core.ModelInfo{ID: id, Thinking: m.Capabilities.Reasoning}
		if m.Capabilities.Reasoning {
			// * only none and high
			info.Efforts = []string{"none", "high"}
		}
		infos = append(infos, info)
	}
	return infos, nil
}
