package copilot

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
	modelsAPI = "https://api.githubcopilot.com/models"
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

func fetchModels(ctx context.Context, client *http.Client, headers map[string]string) (llmrouter.CopilotModels, error) {
	data, status, err := go_pkg_http.GET[llmrouter.CopilotModels](ctx, client, modelsAPI, headers)
	if err != nil {
		return llmrouter.CopilotModels{}, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: %w", err)
	}
	if status != http.StatusOK {
		return llmrouter.CopilotModels{}, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: http %d", status)
	}
	return data, nil
}

func ModelInfos(ctx context.Context, config llmrouter.Config, filter llmrouter.ModelFilter) ([]llmrouter.ModelInfo, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Models: APIKey is required")
	}

	client := &http.Client{Timeout: 10 * time.Second}
	data, err := fetchModels(ctx, client, map[string]string{
		"Authorization":  "Bearer " + config.APIKey,
		"Editor-Version": "vscode/1.95.0",
	})
	if err != nil {
		return nil, err
	}

	infos := make([]llmrouter.ModelInfo, 0, len(data.Data))
	for _, m := range data.Data {
		id := strings.TrimSpace(m.ID)
		if id == "" || !m.ModelPickerEnabled || m.Policy.State == "disabled" {
			continue
		}

		if len(m.SupportedEndpoints) > 0 &&
			!slices.Contains(m.SupportedEndpoints, "/chat/completions") &&
			!slices.Contains(m.SupportedEndpoints, "/responses") {
			continue
		}
		if m.Capabilities.Type != "" && m.Capabilities.Type != "chat" {
			continue
		}
		if filter.TextOnly && !llmrouter.IsTextModel(id) {
			continue
		}
		infos = append(infos, llmrouter.ModelInfo{
			ID:        id,
			Thinking:  len(m.Capabilities.Supports.ReasoningEffort) > 0,
			Efforts:   m.Capabilities.Supports.ReasoningEffort,
			Endpoints: m.SupportedEndpoints,
		})
	}
	return infos, nil
}
