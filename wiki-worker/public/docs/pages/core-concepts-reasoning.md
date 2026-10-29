# Reasoning Levels

How the six `llmrouter.Reasoning` levels map onto each vendor and what happens when a model cannot honor one.

## Levels

The levels map onto each vendor's own mechanism: Claude's thinking budget (4.5 / 4.1) or adaptive-thinking effort (newer models), Gemini's thinking config, OpenAI's reasoning effort.

| Level | String | Alias |
|---|---|---|
| `ReasoningNone` | `none` | — |
| `ReasoningLow` | `low` | `minimal` |
| `ReasoningMedium` | `medium` (default) | — |
| `ReasoningHigh` | `high` | — |
| `ReasoningXHigh` | `xhigh` | `extra` |
| `ReasoningMax` | `max` | `ultra` |

## Clamping to the model's range

A request outside a model's range is not rejected; it is clamped to the nearest legal value and logged at debug level. Every adapter implementing `ReasoningAgent` already does this inside `Send`; call `ClampReasoning` yourself only to learn the effective level up front:

```go
if limited, ok := agent.(llmrouter.ReasoningAgent); ok {
	low, high := limited.ReasoningLimits()
	reasoning = llmrouter.ClampReasoning(reasoning, low, high, "openai", "gpt-5.1")
}
```

`deepseek`, `cloudflare`, and `compat` ignore the level; `claude-code` passes it to the CLI as an effort and sends `none` as `low`.

Clamping beats failing because the caller usually holds one setting shared across models; letting a lower-ceiling model run at its ceiling is closer to what the caller meant than failing the whole request.

## Related

- [Fast Mode](/core-concepts-fast-mode): the other per-request knob
- [Reasoning and Modes](/api-reference-reasoning): signatures
