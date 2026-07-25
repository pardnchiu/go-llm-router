package gemini

import (
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
)

type thinkingMode int

const (
	modeNone thinkingMode = iota
	modeLevel
	modeBudget
)

func limits(model string, thinking *bool) (mode thinkingMode, low, high core.Reasoning) {
	if thinking != nil && !*thinking {
		return modeNone, core.ReasoningNone, core.ReasoningNone
	}

	switch {
	case strings.HasPrefix(model, "gemini-3"), strings.HasSuffix(model, "-latest"):
		return modeLevel, core.ReasoningLow, core.ReasoningHigh
	case strings.HasPrefix(model, "gemini-2.5-"):
		return modeBudget, core.ReasoningNone, core.ReasoningHigh
	}

	if thinking != nil && *thinking && strings.HasPrefix(model, "gemini-") {
		return modeLevel, core.ReasoningLow, core.ReasoningHigh
	}
	return modeNone, core.ReasoningNone, core.ReasoningNone
}

func (a *Agent) ReasoningLimits() (core.Reasoning, core.Reasoning) {
	_, low, high := limits(a.model, a.thinking)
	return low, high
}

var levelName = map[core.Reasoning]string{
	core.ReasoningLow:    "low",
	core.ReasoningMedium: "medium",
	core.ReasoningHigh:   "high",
	core.ReasoningXHigh:  "high",
	core.ReasoningMax:    "high",
}

func thinkingBudget(model string, level core.Reasoning) int {
	switch level {
	case core.ReasoningNone:
		if strings.Contains(model, "2.5-pro") {
			return 128
		}
		return 0
	case core.ReasoningLow:
		return 1024
	case core.ReasoningHigh, core.ReasoningXHigh, core.ReasoningMax:
		return 16384
	}
	return 8192
}

func (a *Agent) applyReasoning(generationConfig map[string]any, reasoning core.Reasoning) {
	mode, low, high := limits(a.model, a.thinking)
	level := core.ClampReasoning(reasoning, low, high, "gemini", a.model)

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
