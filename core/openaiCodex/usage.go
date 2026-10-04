package openaicodex

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

const usageAPI = "https://chatgpt.com/backend-api/wham/usage"

type usageWindow struct {
	UsedPercent float64 `json:"used_percent"`
}

type usageResponse struct {
	RateLimit struct {
		PrimaryWindow   usageWindow `json:"primary_window"`
		SecondaryWindow usageWindow `json:"secondary_window"`
	} `json:"rate_limit"`
}

func Usage(ctx context.Context, config llmrouter.Config) (llmrouter.UsageRemaining, error) {
	if config.APIKey == "" {
		return llmrouter.UsageRemaining{}, fmt.Errorf("Usage: APIKey is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, usageAPI, nil)
	if err != nil {
		return llmrouter.UsageRemaining{}, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+config.APIKey)
	if config.AccountID != "" {
		req.Header.Set("ChatGPT-Account-Id", config.AccountID)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return llmrouter.UsageRemaining{}, fmt.Errorf("http.Do: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, llmrouter.ErrorBodyLimit))
		return llmrouter.UsageRemaining{}, fmt.Errorf("codex usage http %d: %s", resp.StatusCode, raw)
	}

	var usage usageResponse
	if err := json.NewDecoder(io.LimitReader(resp.Body, llmrouter.JSONBodyLimit)).Decode(&usage); err != nil {
		return llmrouter.UsageRemaining{}, fmt.Errorf("json.Decode: %w", err)
	}

	fiveHour := 100 - usage.RateLimit.PrimaryWindow.UsedPercent
	week := 100 - usage.RateLimit.SecondaryWindow.UsedPercent
	return llmrouter.UsageRemaining{FiveHour: &fiveHour, Week: &week}, nil
}
