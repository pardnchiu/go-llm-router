package core

import (
	"log/slog"
	"strings"
)

type Reasoning int

const (
	ReasoningNone Reasoning = iota
	ReasoningLow
	ReasoningMedium
	ReasoningHigh
	ReasoningXHigh
	ReasoningMax
)

const ReasoningDefault = ReasoningMedium

var reasoningNames = [...]string{"none", "low", "medium", "high", "xhigh", "max"}

func (r Reasoning) String() string {
	if r < 0 || int(r) >= len(reasoningNames) {
		return "invalid"
	}
	return reasoningNames[r]
}

var reasoningAliases = map[string]Reasoning{
	"minimal": ReasoningLow,
	"extra":   ReasoningXHigh,
	"ultra":   ReasoningMax,
}

func ParseReasoning(s string) (Reasoning, bool) {
	for i, name := range reasoningNames {
		if name == s {
			return Reasoning(i), true
		}
	}
	if r, ok := reasoningAliases[s]; ok {
		return r, true
	}
	return ReasoningDefault, false
}

func ClampReasoning(r, lo, hi Reasoning, provider, model string) Reasoning {
	out := min(max(r, lo), hi)
	if out != r {
		slog.Debug("reasoning clamped",
			slog.String("provider", provider),
			slog.String("model", model),
			slog.String("requested", r.String()),
			slog.String("effective", out.String()))
	}
	return out
}

type ReasoningAgent interface {
	ReasoningLimits() (min, max Reasoning)
}

func OpenAIEffortRange(model string) (low, high Reasoning) {
	switch {
	case strings.Contains(model, "-chat-latest"):
		return ReasoningNone, ReasoningNone

	case strings.Contains(model, "codex"), openAIMinorAtLeast(model, 6):
		low, high = ReasoningNone, ReasoningMax
	case openAIMinorAtLeast(model, 2):
		low, high = ReasoningNone, ReasoningXHigh
	case openAIMinorAtLeast(model, 1):
		low, high = ReasoningNone, ReasoningHigh
	case strings.HasPrefix(model, "gpt-5"),
		strings.HasPrefix(model, "o1"),
		strings.HasPrefix(model, "o3"),
		strings.HasPrefix(model, "o4"):
		low, high = ReasoningLow, ReasoningHigh
	default:
		return ReasoningNone, ReasoningNone
	}

	if strings.Contains(model, "-pro") {
		low = max(low, ReasoningMedium)
	}
	return low, high
}

func openAIMinorAtLeast(model string, minMinor int) bool {
	rest, ok := strings.CutPrefix(model, "gpt-5.")
	if !ok {
		return false
	}
	end := strings.IndexFunc(rest, func(r rune) bool { return r < '0' || r > '9' })
	if end == -1 {
		end = len(rest)
	}
	if end == 0 {
		return false
	}
	n := 0
	for _, c := range rest[:end] {
		n = n*10 + int(c-'0')
	}
	return n >= minMinor
}

type ModelFilter struct {
	TextOnly bool
}

var nonTextMarkers = []string{
	"-image", "-tts", "-audio", "-video", "-live",
	"imagine-", "lyria-", "nano-banana", "veo-", "imagen-",
	"-robotics-", "computer-use", "deep-research", "antigravity", "gemini-omni",
	"multi-agent",
	"image-", "realtime", "sora-", "embedding-", "tts-", "whisper-",
	"babbage-", "davinci-", "search-", "transcribe", "moderation-",
}

func IsTextModel(id string) bool {
	for _, marker := range nonTextMarkers {
		if strings.Contains(id, marker) {
			return false
		}
	}
	return true
}
