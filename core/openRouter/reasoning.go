package openrouter

import (
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

func limits(model string) (low, high llmrouter.Reasoning) {
	if vendor, _, _ := strings.Cut(model, "/"); vendor == "deepseek" {
		return llmrouter.ReasoningNone, llmrouter.ReasoningNone
	}
	return llmrouter.ReasoningNone, llmrouter.ReasoningMax
}

func (a *Agent) ReasoningLimits() (llmrouter.Reasoning, llmrouter.Reasoning) {
	return limits(a.model)
}

var effortName = map[llmrouter.Reasoning]string{
	llmrouter.ReasoningLow:    "low",
	llmrouter.ReasoningMedium: "medium",
	llmrouter.ReasoningHigh:   "high",
	llmrouter.ReasoningXHigh:  "xhigh",
	llmrouter.ReasoningMax:    "max",
}

func (a *Agent) effort(reasoning llmrouter.Reasoning) (string, bool) {
	low, high := limits(a.model)
	if high == llmrouter.ReasoningNone {
		return "", false
	}
	level := llmrouter.ClampReasoning(reasoning, low, high, "openrouter", a.model)
	if level == llmrouter.ReasoningNone {
		return "", false
	}
	return effortName[level], true
}
