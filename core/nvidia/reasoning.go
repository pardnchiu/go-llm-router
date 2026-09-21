package nvidia

import (
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

func limits(model string) (low, high llmrouter.Reasoning) {
	if !strings.Contains(model, "gpt-oss") {
		return llmrouter.ReasoningNone, llmrouter.ReasoningNone
	}
	return llmrouter.ReasoningLow, llmrouter.ReasoningHigh
}

func (a *Agent) ReasoningLimits() (llmrouter.Reasoning, llmrouter.Reasoning) {
	return limits(a.model)
}

var effortName = map[llmrouter.Reasoning]string{
	llmrouter.ReasoningLow:    "low",
	llmrouter.ReasoningMedium: "medium",
	llmrouter.ReasoningHigh:   "high",
	llmrouter.ReasoningXHigh:  "high",
	llmrouter.ReasoningMax:    "high",
}

func (a *Agent) effort(reasoning llmrouter.Reasoning) (string, bool) {
	low, high := limits(a.model)
	if high == llmrouter.ReasoningNone {
		return "", false
	}

	level := llmrouter.ClampReasoning(reasoning, low, high, "nvidia", a.model)
	return effortName[level], true
}
