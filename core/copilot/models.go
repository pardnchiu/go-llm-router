package copilot

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
	modelsAPI = "https://api.githubcopilot.com/models"
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
	data, status, err := go_pkg_http.GET[core.CopilotModels](ctx, client, modelsAPI, map[string]string{
		"Authorization":  "Bearer " + config.APIKey,
		"Editor-Version": "vscode/1.95.0",
	})
	if err != nil {
		return nil, fmt.Errorf("go_pkg_http.GET: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("go_pkg_http.GET: http %d", status)
	}

	infos := make([]core.ModelInfo, 0, len(data.Data))
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
		if filter.TextOnly && !core.IsTextModel(id) {
			continue
		}
		infos = append(infos, core.ModelInfo{
			ID:        id,
			Thinking:  len(m.Capabilities.Supports.ReasoningEffort) > 0,
			Efforts:   m.Capabilities.Supports.ReasoningEffort,
			Endpoints: m.SupportedEndpoints,
		})
	}
	return infos, nil
}
