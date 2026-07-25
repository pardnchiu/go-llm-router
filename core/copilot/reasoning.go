package copilot

import (
	"slices"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
)

var effortLevel = map[string]core.Reasoning{
	"none":    core.ReasoningNone,
	"minimal": core.ReasoningLow,
	"low":     core.ReasoningLow,
	"medium":  core.ReasoningMedium,
	"high":    core.ReasoningHigh,
	"xhigh":   core.ReasoningXHigh,
	"max":     core.ReasoningMax,
}

var levelEffort = map[core.Reasoning]string{
	core.ReasoningNone:   "none",
	core.ReasoningLow:    "low",
	core.ReasoningMedium: "medium",
	core.ReasoningHigh:   "high",
	core.ReasoningXHigh:  "xhigh",
	core.ReasoningMax:    "max",
}

func limits(model string, efforts []string) (low, high core.Reasoning) {
	if len(efforts) > 0 {
		low, high = core.ReasoningMax, core.ReasoningNone
		for _, e := range efforts {
			level, ok := effortLevel[e]
			if !ok {
				continue
			}
			low = min(low, level)
			high = max(high, level)
		}
		if high < low {
			return core.ReasoningNone, core.ReasoningNone
		}
		return low, high
	}

	if !strings.HasPrefix(model, "gpt-5") && !strings.Contains(model, "codex") {
		return core.ReasoningNone, core.ReasoningNone
	}
	return core.ReasoningNone, core.ReasoningXHigh
}

func (a *Agent) ReasoningLimits() (core.Reasoning, core.Reasoning) {
	return limits(a.model, a.efforts)
}

func (a *Agent) effort(reasoning core.Reasoning) (string, bool) {
	low, high := limits(a.model, a.efforts)
	if low == high && low == core.ReasoningNone {
		return "", false
	}

	level := core.ClampReasoning(reasoning, low, high, "copilot", a.model)
	name := levelEffort[level]
	if len(a.efforts) > 0 && !slices.Contains(a.efforts, name) {
		for l := level; l <= core.ReasoningMax; l++ {
			if candidate := levelEffort[l]; slices.Contains(a.efforts, candidate) {
				name = candidate
				break
			}
		}
	}

	if name == "none" || name == "" {
		return "", false
	}
	return name, true
}
