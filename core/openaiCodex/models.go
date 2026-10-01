package openaicodex

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	go_pkg_http "github.com/pardnchiu/go-pkg/http"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

const (
	versionAPI = "https://registry.npmjs.org/-/package/@openai/codex/dist-tags"
	modelsAPI  = "https://chatgpt.com/backend-api/codex/models"
)

type distTags struct {
	Latest string `json:"latest"`
}

type codexModels struct {
	Models []struct {
		Slug       string `json:"slug"`
		Visibility string `json:"visibility"`
	} `json:"models"`
}

func Models(ctx context.Context, config llmrouter.Config, filter llmrouter.ModelFilter) ([]string, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("Models: APIKey is required")
	}

	client := &http.Client{Timeout: 10 * time.Second}

	tags, status, err := go_pkg_http.GET[distTags](ctx, client, versionAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: http %d", status)
	}
	version := strings.TrimSpace(tags.Latest)
	if version == "" {
		return nil, fmt.Errorf("Models: codex latest version is empty")
	}

	headers := map[string]string{
		"Authorization": "Bearer " + config.APIKey,
	}
	if config.AccountID != "" {
		headers["ChatGPT-Account-Id"] = config.AccountID
	}

	data, status, err := go_pkg_http.GET[codexModels](ctx, client, modelsAPI+"?client_version="+url.QueryEscape(version), headers)
	if err != nil {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("github.com/pardnchiu/go-pkg/http: GET: http %d", status)
	}

	ids := make([]string, 0, len(data.Models))
	for _, m := range data.Models {
		id := strings.TrimSpace(m.Slug)
		if id == "" || m.Visibility != "list" {
			continue
		}
		if filter.TextOnly && !llmrouter.IsTextModel(id) {
			continue
		}
		ids = append(ids, id)
	}
	return ids, nil
}
