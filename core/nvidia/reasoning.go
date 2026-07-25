package nvidia

import (
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
)

func limits(model string) (low, high core.Reasoning) {
	if !strings.Contains(model, "gpt-oss") {
		return core.ReasoningNone, core.ReasoningNone
	}
	return core.ReasoningLow, core.ReasoningHigh
}

func (a *Agent) ReasoningLimits() (core.Reasoning, core.Reasoning) {
	return limits(a.model)
}

var effortName = map[core.Reasoning]string{
	core.ReasoningLow:    "low",
	core.ReasoningMedium: "medium",
	core.ReasoningHigh:   "high",
	core.ReasoningXHigh:  "high",
	core.ReasoningMax:    "high",
}

func (a *Agent) effort(reasoning core.Reasoning) (string, bool) {
	low, high := limits(a.model)
	if high == core.ReasoningNone {
		return "", false
	}

	level := core.ClampReasoning(reasoning, low, high, "nvidia", a.model)
	return effortName[level], true
}
