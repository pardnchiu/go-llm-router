# Reasoning and Modes

Constants and functions that set reasoning effort and service tier per request.

## Reasoning

| Symbol | Signature |
|---|---|
| `Reasoning` constants | `ReasoningNone`, `ReasoningLow`, `ReasoningMedium`, `ReasoningHigh`, `ReasoningXHigh`, `ReasoningMax`; `ReasoningDefault = ReasoningMedium` |
| `(Reasoning).String` | `func() string` |
| `ParseReasoning` | `func(s string) (Reasoning, bool)`; aliases `minimal` / `extra` / `ultra` |
| `ClampReasoning` | `func(r, lo, hi Reasoning, provider, model string) Reasoning` |
| `OpenAIEffortRange` | `func(model string) (low, high Reasoning)` |

## Modes

| Symbol | Signature |
|---|---|
| `Mode` constants | `ModeDefault`, `ModeFast` |
| `(Mode).String` | `func() string` |
| `ParseMode` | `func(s string) (Mode, bool)` |
| `SupportFast` | `func(providerName, model string) bool` |
| `WarnFastDowngrade` | `func(providerName, model, tier string)` |

## Related

- [Reasoning Levels](/core-concepts-reasoning)
- [Fast Mode](/core-concepts-fast-mode)
