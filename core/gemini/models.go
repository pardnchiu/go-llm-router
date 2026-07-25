package gemini

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/pardnchiu/go-llm-router/core"
	go_pkg_http "github.com/pardnchiu/go-pkg/http"
)

const (
	modelsAPI = "https://generativelanguage.googleapis.com/v1beta/models"
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
	data, status, err := go_pkg_http.GET[core.GeminiModels](ctx, client, modelsAPI+"?key="+config.APIKey, nil)
	if err != nil {
		return nil, fmt.Errorf("go_pkg_http.GET: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("go_pkg_http.GET: http %d", status)
	}

	infos := make([]core.ModelInfo, 0, len(data.Models))
	for _, m := range data.Models {
		name := strings.TrimPrefix(m.Name, "models/")
		if name == "" || !slices.Contains(m.SupportedGenerationMethods, "generateContent") {
			continue
		}
		if filter.TextOnly && !core.IsTextModel(name) {
			continue
		}
		infos = append(infos, core.ModelInfo{ID: name, Thinking: m.Thinking})
	}
	return infos, nil
}
