package grokoauth

import (
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
)

func limits(model string) (low, high core.Reasoning) {
	switch {
	case strings.HasPrefix(model, "grok-4.20"), strings.HasPrefix(model, "grok-build"):
		return core.ReasoningNone, core.ReasoningNone
	case strings.HasPrefix(model, "grok-4.5"):
		return core.ReasoningLow, core.ReasoningXHigh
	}
	return core.ReasoningNone, core.ReasoningXHigh
}

func (a *Agent) ReasoningLimits() (core.Reasoning, core.Reasoning) {
	return limits(a.model)
}

var effortName = map[core.Reasoning]string{
	core.ReasoningLow:    "low",
	core.ReasoningMedium: "medium",
	core.ReasoningHigh:   "high",
	core.ReasoningXHigh:  "xhigh",
	core.ReasoningMax:    "xhigh",
}

func (a *Agent) effort(reasoning core.Reasoning) (string, bool) {
	low, high := limits(a.model)
	if low == high && low == core.ReasoningNone {
		return "", false
	}

	level := core.ClampReasoning(reasoning, low, high, "grok-oauth", a.model)
	if level == core.ReasoningNone {
		return "none", true
	}
	return effortName[level], true
}
