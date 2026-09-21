package ollamacloud

import (
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

func limits(model string) (low, high llmrouter.Reasoning) {
	if strings.HasPrefix(model, "gpt-oss") {
		return llmrouter.ReasoningLow, llmrouter.ReasoningHigh
	}
	return llmrouter.ReasoningNone, llmrouter.ReasoningHigh
}

func (a *Agent) ReasoningLimits() (llmrouter.Reasoning, llmrouter.Reasoning) {
	return limits(a.model)
}

var effortName = map[llmrouter.Reasoning]string{
	llmrouter.ReasoningNone:   "none",
	llmrouter.ReasoningLow:    "low",
	llmrouter.ReasoningMedium: "medium",
	llmrouter.ReasoningHigh:   "high",
}

func (a *Agent) effort(reasoning llmrouter.Reasoning) string {
	low, high := limits(a.model)
	return effortName[llmrouter.ClampReasoning(reasoning, low, high, label, a.model)]
}
