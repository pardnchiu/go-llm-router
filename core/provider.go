package core

import (
	"net/http"
	"strings"
	"time"
)

func SupportTemperature(providerName, model string) bool {
	switch providerName {
	case "openai", "copilot", "codex":
		if strings.HasPrefix(model, "gpt-5") {
			return false
		}
	case "deepseek":
		if model == "deepseek-reasoner" {
			return false
		}
	case "claude":
		return false
	case "gemini":
		if strings.Contains(model, "-preview") {
			return false
		}
	}
	return true
}

func ResponsesAPI(providerName, model string) bool {
	switch providerName {
	case "openai":
		return strings.Contains(model, "codex") || openAIMinorAtLeast(model, 4) || strings.HasSuffix(model, "-pro")
	case "copilot":
		return strings.Contains(model, "-codex") || openAIMinorAtLeast(model, 4)
	}
	return false
}

func NewHTTPClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Minute}
}
