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

	case strings.Contains(model, "codex"), openAIVersionAtLeast(model, 5, 6):
		low, high = ReasoningNone, ReasoningMax
	case openAIVersionAtLeast(model, 5, 2):
		low, high = ReasoningNone, ReasoningXHigh
	case openAIVersionAtLeast(model, 5, 1):
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

func openAIVersionAtLeast(model string, minMajor, minMinor int) bool {
	major, minor, ok := openAIVersion(model)
	if !ok {
		return false
	}
	if major != minMajor {
		return major > minMajor
	}
	return minor >= minMinor
}

func openAIVersion(model string) (int, int, bool) {
	rest, ok := strings.CutPrefix(model, "gpt-")
	if !ok {
		return 0, 0, false
	}

	majorText, minorText, hasMinor := strings.Cut(rest, ".")
	major, ok := leadingInt(majorText)
	if !ok {
		return 0, 0, false
	}
	if !hasMinor {
		return major, 0, true
	}

	minor, ok := leadingInt(minorText)
	if !ok {
		return major, 0, true
	}
	return major, minor, true
}

func leadingInt(text string) (int, bool) {
	end := strings.IndexFunc(text, func(r rune) bool { return r < '0' || r > '9' })
	if end == -1 {
		end = len(text)
	}
	if end == 0 {
		return 0, false
	}

	n := 0
	for _, c := range text[:end] {
		n = n*10 + int(c-'0')
	}
	return n, true
}

type ModelFilter struct {
	TextOnly bool
	STTOnly  bool
	TTSOnly  bool
}

var nonTextMarkers = []string{
	"-image", "-tts", "-audio", "-video", "-live",
	"imagine-", "lyria-", "nano-banana", "veo-", "imagen-",
	"-robotics-", "computer-use", "deep-research", "antigravity", "gemini-omni",
	"multi-agent",
	"image-", "realtime", "sora-", "embedding-", "tts-", "whisper-",
	"babbage-", "davinci-", "search-", "transcribe", "moderation-",
}

var sttMarkers = []string{
	"transcribe", "whisper-",
}

var ttsMarkers = []string{
	"-tts", "tts-",
}

var nonHTTPMarkers = []string{
	"-live",
}

func containsAny(id string, markers []string) bool {
	for _, marker := range markers {
		if strings.Contains(id, marker) {
			return true
		}
	}
	return false
}

func IsTextModel(id string) bool {
	return !containsAny(id, nonTextMarkers)
}

func IsSTTModel(id string) bool {
	return !containsAny(id, nonHTTPMarkers) && containsAny(id, sttMarkers)
}

func IsTTSModel(id string) bool {
	return !containsAny(id, nonHTTPMarkers) && containsAny(id, ttsMarkers)
}

func MatchModelFilter(id string, filter ModelFilter) bool {
	if filter.TextOnly && !IsTextModel(id) {
		return false
	}
	if filter.STTOnly && !IsSTTModel(id) {
		return false
	}
	if filter.TTSOnly && !IsTTSModel(id) {
		return false
	}
	return true
}
