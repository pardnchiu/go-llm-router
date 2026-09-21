package gemini

import (
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

type thinkingMode int

const (
	modeNone thinkingMode = iota
	modeLevel
	modeBudget
)

func limits(model string, thinking *bool) (mode thinkingMode, low, high llmrouter.Reasoning) {
	if thinking != nil && !*thinking {
		return modeNone, llmrouter.ReasoningNone, llmrouter.ReasoningNone
	}

	switch {
	case strings.HasPrefix(model, "gemini-3"), strings.HasSuffix(model, "-latest"):
		return modeLevel, llmrouter.ReasoningLow, llmrouter.ReasoningHigh
	case strings.HasPrefix(model, "gemini-2.5-"):
		return modeBudget, llmrouter.ReasoningNone, llmrouter.ReasoningHigh
	}

	if thinking != nil && *thinking && strings.HasPrefix(model, "gemini-") {
		return modeLevel, llmrouter.ReasoningLow, llmrouter.ReasoningHigh
	}
	return modeNone, llmrouter.ReasoningNone, llmrouter.ReasoningNone
}

func (a *Agent) ReasoningLimits() (llmrouter.Reasoning, llmrouter.Reasoning) {
	_, low, high := limits(a.model, a.thinking)
	return low, high
}

var levelName = map[llmrouter.Reasoning]string{
	llmrouter.ReasoningLow:    "low",
	llmrouter.ReasoningMedium: "medium",
	llmrouter.ReasoningHigh:   "high",
	llmrouter.ReasoningXHigh:  "high",
	llmrouter.ReasoningMax:    "high",
}

func thinkingBudget(model string, level llmrouter.Reasoning) int {
	switch level {
	case llmrouter.ReasoningNone:
		if strings.Contains(model, "2.5-pro") {
			return 128
		}
		return 0
	case llmrouter.ReasoningLow:
		return 1024
	case llmrouter.ReasoningHigh, llmrouter.ReasoningXHigh, llmrouter.ReasoningMax:
		return 16384
	}
	return 8192
}

func (a *Agent) applyReasoning(generationConfig map[string]any, reasoning llmrouter.Reasoning) {
	mode, low, high := limits(a.model, a.thinking)
	level := llmrouter.ClampReasoning(reasoning, low, high, "gemini", a.model)

	switch mode {
	case modeLevel:
		generationConfig["thinkingConfig"] = map[string]any{
			"thinkingLevel":   levelName[level],
			"includeThoughts": true,
		}
	case modeBudget:
		generationConfig["temperature"] = 0.2
		budget := thinkingBudget(a.model, level)
		thinking := map[string]any{"thinkingBudget": budget}
		if budget > 0 {
			thinking["includeThoughts"] = true
		}
		generationConfig["thinkingConfig"] = thinking
	default:
		generationConfig["temperature"] = 0.2
	}
}
