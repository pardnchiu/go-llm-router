package grokoauth

import (
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

func limits(model string) (low, high llmrouter.Reasoning) {
	switch {
	case strings.HasPrefix(model, "grok-4.20"), strings.HasPrefix(model, "grok-build"):
		return llmrouter.ReasoningNone, llmrouter.ReasoningNone
	case strings.HasPrefix(model, "grok-4.5"):
		return llmrouter.ReasoningLow, llmrouter.ReasoningXHigh
	}
	return llmrouter.ReasoningNone, llmrouter.ReasoningXHigh
}

func (a *Agent) ReasoningLimits() (llmrouter.Reasoning, llmrouter.Reasoning) {
	return limits(a.model)
}

var effortName = map[llmrouter.Reasoning]string{
	llmrouter.ReasoningLow:    "low",
	llmrouter.ReasoningMedium: "medium",
	llmrouter.ReasoningHigh:   "high",
	llmrouter.ReasoningXHigh:  "xhigh",
	llmrouter.ReasoningMax:    "xhigh",
}

func (a *Agent) effort(reasoning llmrouter.Reasoning) (string, bool) {
	low, high := limits(a.model)
	if low == high && low == llmrouter.ReasoningNone {
		return "", false
	}

	level := llmrouter.ClampReasoning(reasoning, low, high, "grok-oauth", a.model)
	if level == llmrouter.ReasoningNone {
		return "none", true
	}
	return effortName[level], true
}
