package ollamacloud

import (
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
)

func limits(model string) (low, high core.Reasoning) {
	if strings.HasPrefix(model, "gpt-oss") {
		return core.ReasoningLow, core.ReasoningHigh
	}
	return core.ReasoningNone, core.ReasoningHigh
}

func (a *Agent) ReasoningLimits() (core.Reasoning, core.Reasoning) {
	return limits(a.model)
}

var effortName = map[core.Reasoning]string{
	core.ReasoningNone:   "none",
	core.ReasoningLow:    "low",
	core.ReasoningMedium: "medium",
	core.ReasoningHigh:   "high",
}

func (a *Agent) effort(reasoning core.Reasoning) string {
	low, high := limits(a.model)
	return effortName[core.ClampReasoning(reasoning, low, high, label, a.model)]
}
