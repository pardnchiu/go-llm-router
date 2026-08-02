package core

import (
	"log/slog"
	"slices"
	"strings"
)

type Mode int

const (
	ModeDefault Mode = iota
	ModeFast
)

var modeNames = [...]string{"default", "fast"}

func (m Mode) String() string {
	if m < 0 || int(m) >= len(modeNames) {
		return "invalid"
	}
	return modeNames[m]
}

func ParseMode(s string) (Mode, bool) {
	for i, name := range modeNames {
		if name == s {
			return Mode(i), true
		}
	}
	return ModeDefault, false
}

func SupportFast(providerName, model string) bool {
	return fastSupported(providerName, model)
}

func WarnFastDowngrade(providerName, model, tier string) {
	switch tier {
	case "", "fast", "priority":
		return
	}
	slog.Warn("fast mode downgraded",
		slog.String("provider", providerName),
		slog.String("model", model),
		slog.String("tier", tier))
}

var geminiFastPrefixes = []string{
	"gemini-3.5-flash",
	"gemini-3.1-flash-lite",
	"gemini-3.1-pro",
	"gemini-3-flash",
	"gemini-2.5-pro",
	"gemini-2.5-flash",
}

func fastSupported(providerName, model string) bool {
	switch providerName {
	case "openai":
		if strings.Contains(model, "-pro") || strings.Contains(model, "-nano") {
			return false
		}
		if strings.Contains(model, "codex") {
			return openAIVersionAtLeast(model, 5, 3)
		}
		return openAIVersionAtLeast(model, 5, 4)

	case "claude":
		return strings.Contains(model, "opus-5") || strings.Contains(model, "opus-4-8")

	case "grok":
		return !strings.Contains(model, "imagine")

	case "gemini":
		return slices.ContainsFunc(geminiFastPrefixes, func(prefix string) bool {
			return strings.HasPrefix(model, prefix)
		})

	case "openrouter":
		vendor, id, ok := strings.Cut(model, "/")
		if !ok {
			return false
		}
		switch vendor {
		case "openai":
			return fastSupported("openai", id)
		case "google":
			return fastSupported("gemini", id)
		case "x-ai":
			return fastSupported("grok", id)
		}
	}
	return false
}
