# Agent Contract

`Agent` is the only interface every provider implements; every other capability is reached by type assertion.

## The Agent interface

```go
type Agent interface {
	Name() string
	Send(ctx context.Context, messages []Message, toolDefs []Tool, reasoning Reasoning, mode Mode) (*Output, int, error)
}
```

`Send` returns the upstream HTTP status as its second value, so callers branch on a status code rather than on error strings.

## Optional capabilities

An agent without a capability fails the assertion.

| Optional interface | Method | Implemented by |
|---|---|---|
| `StreamAgent` | `SendStream` | every provider except `claude-code`; `cloudflare` always returns `ErrStreamUnsupported` |
| `ReasoningAgent` | `ReasoningLimits` | every provider except `deepseek`, `cloudflare`, `compat`, and `claude-code` |
| `ImageAgent` | `GenerateImage` | `openai`, `codex`, `gemini`, `grok`, `grok-oauth`, `openrouter` |
| `STTAgent` | `Transcribe` | `openai`, `gemini`, `openrouter` |
| `TTSAgent` | `Speak` | `openai`, `gemini`, `openrouter` |

## Why interfaces instead of flags

The split exists so capability differences live in the type system: one assertion answers "can this agent do it", and no lookup table of vendor features has to be maintained by hand.

## Related

- [Model Routing](/core-concepts-routing): how `router.New` picks the adapter
- [Interfaces and Router](/api-reference): full interface signatures
