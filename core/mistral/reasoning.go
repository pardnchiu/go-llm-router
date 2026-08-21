package mistral

import (
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
)

func limits(model string, thinking *bool) (low, high core.Reasoning) {
	if thinking != nil {
		if !*thinking {
			return core.ReasoningNone, core.ReasoningNone
		}
		return core.ReasoningNone, core.ReasoningHigh
	}
	if strings.HasPrefix(model, "magistral") {
		return core.ReasoningNone, core.ReasoningHigh
	}
	return core.ReasoningNone, core.ReasoningNone
}

func (a *Agent) ReasoningLimits() (core.Reasoning, core.Reasoning) {
	return limits(a.model, a.thinking)
}

func (a *Agent) effort(reasoning core.Reasoning) (string, bool) {
	low, high := limits(a.model, a.thinking)
	if high == core.ReasoningNone {
		return "", false
	}
	if core.ClampReasoning(reasoning, low, high, "mistral", a.model) == core.ReasoningNone {
		return "none", true
	}
	return "high", true
}
