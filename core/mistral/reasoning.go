package mistral

import (
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

func limits(model string, thinking *bool) (low, high llmrouter.Reasoning) {
	if thinking != nil {
		if !*thinking {
			return llmrouter.ReasoningNone, llmrouter.ReasoningNone
		}
		return llmrouter.ReasoningNone, llmrouter.ReasoningHigh
	}
	if strings.HasPrefix(model, "magistral") {
		return llmrouter.ReasoningNone, llmrouter.ReasoningHigh
	}
	return llmrouter.ReasoningNone, llmrouter.ReasoningNone
}

func (a *Agent) ReasoningLimits() (llmrouter.Reasoning, llmrouter.Reasoning) {
	return limits(a.model, a.thinking)
}

func (a *Agent) effort(reasoning llmrouter.Reasoning) (string, bool) {
	low, high := limits(a.model, a.thinking)
	if high == llmrouter.ReasoningNone {
		return "", false
	}
	if llmrouter.ClampReasoning(reasoning, low, high, "mistral", a.model) == llmrouter.ReasoningNone {
		return "none", true
	}
	return "high", true
}
