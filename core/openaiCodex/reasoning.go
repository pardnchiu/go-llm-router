package openaicodex

import (
	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

func limits(model string) (low, high llmrouter.Reasoning) {
	return llmrouter.OpenAIEffortRange(model)
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
	if low == high && low == llmrouter.ReasoningNone {
		return "", false
	}
	level := llmrouter.ClampReasoning(reasoning, low, high, "codex", a.model)
	if level == llmrouter.ReasoningNone {
		return "", false
	}
	return effortName[level], true
}
