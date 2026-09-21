package openai

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
	lo, hi := limits(a.model)
	if lo == hi && lo == llmrouter.ReasoningNone {
		return "", false
	}
	level := llmrouter.ClampReasoning(reasoning, lo, hi, "openai", a.model)
	if level == llmrouter.ReasoningNone {
		return "", false
	}
	return effortName[level], true
}
