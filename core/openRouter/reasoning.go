package openrouter

import (
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
)

func limits(model string) (low, high core.Reasoning) {
	if vendor, _, _ := strings.Cut(model, "/"); vendor == "deepseek" {
		return core.ReasoningNone, core.ReasoningNone
	}
	return core.ReasoningNone, core.ReasoningMax
}

func (a *Agent) ReasoningLimits() (core.Reasoning, core.Reasoning) {
	return limits(a.model)
}

var effortName = map[core.Reasoning]string{
	core.ReasoningLow:    "low",
	core.ReasoningMedium: "medium",
	core.ReasoningHigh:   "high",
	core.ReasoningXHigh:  "xhigh",
	core.ReasoningMax:    "max",
}

func (a *Agent) effort(reasoning core.Reasoning) (string, bool) {
	low, high := limits(a.model)
	if high == core.ReasoningNone {
		return "", false
	}
	level := core.ClampReasoning(reasoning, low, high, "openrouter", a.model)
	if level == core.ReasoningNone {
		return "", false
	}
	return effortName[level], true
}
