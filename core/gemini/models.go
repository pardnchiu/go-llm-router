package gemini

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	modelsAPI = "https://generativelanguage.googleapis.com/v1beta/models"
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
	data, status, err := go_pkg_http.GET[llmrouter.GeminiModels](ctx, client, modelsAPI+"?key="+config.APIKey, nil)
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: http %d", status)
	}

	infos := make([]llmrouter.ModelInfo, 0, len(data.Models))
	for _, m := range data.Models {
		name := strings.TrimPrefix(m.Name, "models/")
		if name == "" || !slices.Contains(m.SupportedGenerationMethods, "generateContent") {
			continue
		}
		if !llmrouter.MatchModelFilter(name, filter) {
			continue
		}
		infos = append(infos, llmrouter.ModelInfo{ID: name, Thinking: m.Thinking})
	}
	return infos, nil
}
