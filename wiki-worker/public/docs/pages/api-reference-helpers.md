# Helpers

Standalone helper functions outside the Agent interfaces.

## HTTP and model checks

| Symbol | Signature | Description |
|---|---|---|
| `NewHTTPClient` | `func() *http.Client` | 10-minute timeout |
| `SupportTemperature` | `func(providerName, model string) bool` | Claude, the `gpt-5` family, `deepseek-reasoner`, and Gemini previews reject it |
| `ResponsesAPI` | `func(providerName, model string) bool` | Chooses Responses over Chat Completions |
| `summary.VoiceReply` (`core/gemini/summary`) | `func(ctx context.Context, text string) (string, error)` | Compresses text over 320 characters into a speakable reply via Gemini |

## Session

| Symbol | Signature | Description |
|---|---|---|
| `WithSessionID` | `func(ctx context.Context, id string) context.Context` | Attaches a trimmed session ID; an empty ID returns `ctx` unchanged |
| `SessionID` | `func(ctx context.Context) string` | The attached ID, or `""` |
| `SessionUUID` | `func(ctx context.Context) string` | SHA-256 of the ID formatted as a UUID; `""` when no ID is attached |
| `WithCacheTTL` | `func(ctx context.Context, ttl string) context.Context` | Attaches a trimmed prompt-cache TTL such as `"1h"`; an empty value returns `ctx` unchanged |
| `CacheTTL` | `func(ctx context.Context) string` | The attached TTL, or `""` |

## Related

- [Session Cache Keys](/core-concepts-sessions): which providers use the key and the TTL
