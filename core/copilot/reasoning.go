package copilot

import (
	"slices"
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

var effortLevel = map[string]llmrouter.Reasoning{
	"none":    llmrouter.ReasoningNone,
	"minimal": llmrouter.ReasoningLow,
	"low":     llmrouter.ReasoningLow,
	"medium":  llmrouter.ReasoningMedium,
	"high":    llmrouter.ReasoningHigh,
	"xhigh":   llmrouter.ReasoningXHigh,
	"max":     llmrouter.ReasoningMax,
}

var levelEffort = map[llmrouter.Reasoning]string{
	llmrouter.ReasoningNone:   "none",
	llmrouter.ReasoningLow:    "low",
	llmrouter.ReasoningMedium: "medium",
	llmrouter.ReasoningHigh:   "high",
	llmrouter.ReasoningXHigh:  "xhigh",
	llmrouter.ReasoningMax:    "max",
}

func limits(model string, efforts []string) (low, high llmrouter.Reasoning) {
	if len(efforts) > 0 {
		low, high = llmrouter.ReasoningMax, llmrouter.ReasoningNone
		for _, e := range efforts {
			level, ok := effortLevel[e]
			if !ok {
				continue
			}
			low = min(low, level)
			high = max(high, level)
		}
		if high < low {
			return llmrouter.ReasoningNone, llmrouter.ReasoningNone
		}
		return low, high
	}

	if !strings.HasPrefix(model, "gpt-5") && !strings.Contains(model, "codex") {
		return llmrouter.ReasoningNone, llmrouter.ReasoningNone
	}
	return llmrouter.ReasoningNone, llmrouter.ReasoningXHigh
}

func (a *Agent) ReasoningLimits() (llmrouter.Reasoning, llmrouter.Reasoning) {
	return limits(a.model, a.efforts)
}

func (a *Agent) effort(reasoning llmrouter.Reasoning) (string, bool) {
	low, high := limits(a.model, a.efforts)
	if low == high && low == llmrouter.ReasoningNone {
		return "", false
	}

	level := llmrouter.ClampReasoning(reasoning, low, high, "copilot", a.model)
	name := levelEffort[level]
	if len(a.efforts) > 0 && !slices.Contains(a.efforts, name) {
		for l := level; l <= llmrouter.ReasoningMax; l++ {
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
