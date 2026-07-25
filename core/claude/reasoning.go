package claude

import (
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
)

type thinkingMode int

const (
	modeBudget thinkingMode = iota
	mode46
	mode47
)

func limits(model string) (mode thinkingMode, low, high core.Reasoning) {
	if strings.Contains(model, "-4-5") || strings.Contains(model, "-4-1") {
		return modeBudget, core.ReasoningNone, core.ReasoningHigh
	}
	mode = mode47
	if strings.Contains(model, "opus-4-6") || strings.Contains(model, "sonnet-4-6") {
		mode = mode46
	}
	return mode, core.ReasoningNone, core.ReasoningMax
}

func (a *Agent) ReasoningLimits() (core.Reasoning, core.Reasoning) {
	_, low, high := limits(a.model)
	return low, high
}

var effortName = map[core.Reasoning]string{
	core.ReasoningLow:    "low",
	core.ReasoningMedium: "medium",
	core.ReasoningHigh:   "high",
	core.ReasoningXHigh:  "xhigh",
	core.ReasoningMax:    "max",
}

var thinkingBudget = map[core.Reasoning]int{
	core.ReasoningLow:    5000,
	core.ReasoningMedium: 10000,
	core.ReasoningHigh:   32000,
}

func adaptiveThinking() map[string]any {
	return map[string]any{"type": "adaptive", "display": "summarized"}
}

func (a *Agent) applyReasoning(body map[string]any, reasoning core.Reasoning) {
	mode, low, high := limits(a.model)
	level := core.ClampReasoning(reasoning, low, high, "claude", a.model)

	if level == core.ReasoningNone {
		switch {
		case mode == modeBudget:
			body["temperature"] = 0.2
		case strings.Contains(a.model, "fable"), strings.Contains(a.model, "mythos"):
		default:
			body["thinking"] = map[string]any{"type": "disabled"}
		}
		return
	}

	switch mode {
	case modeBudget:
		budget, ok := thinkingBudget[level]
		if !ok {
			budget = thinkingBudget[core.ReasoningMedium]
		}
		if ceiling := a.maxOutputTokens() / 2; budget > ceiling {
			budget = ceiling
		}
		body["thinking"] = map[string]any{
			"type":          "enabled",
			"budget_tokens": budget,
		}
	case mode46:
		body["thinking"] = adaptiveThinking()
		effor := effortName[level]
		// * xhigh only support in 4.7
		if effor == "xhigh" {
			effor = "max"
		}
		body["output_config"] = map[string]any{"effort": effor}
	default:
		body["thinking"] = adaptiveThinking()
		body["output_config"] = map[string]any{"effort": effortName[level]}
	}
}
