# Helpers

Standalone helper functions outside the Agent interfaces.

## HTTP and model checks

| Symbol | Signature | Description |
|---|---|---|
| `NewHTTPClient` | `func() *http.Client` | `TimeoutHTTPLong` (15-minute) overall timeout |
| `NewStreamHTTPClient` | `func() *http.Client` | `TimeoutHTTPLong` overall timeout plus a `TimeoutHTTPShort` (15-second) response-header timeout on a clone of `http.DefaultTransport`; used by `codex`, `grok-oauth`, and the Codex / Grok OAuth requests |
| `SupportTemperature` | `func(providerName, model string) bool` | Claude, the `gpt-5` family, `deepseek-reasoner`, and Gemini previews reject it |
| `ResponsesAPI` | `func(providerName, model string) bool` | Chooses Responses over Chat Completions; for `copilot` it applies only when the model's `supported_endpoints` lists neither `/responses` nor `/chat/completions` |
| `summary.VoiceReply` (`core/gemini/summary`) | `func(ctx context.Context, text string) (string, error)` | Compresses text over 320 characters into a speakable reply via Gemini |

## Timing constants

| Symbol | Value | Used for |
|---|---|---|
| `TimeoutHTTPShort` | 15 s | Response-header timeout, model listing / usage queries, OAuth callback server read / write, Copilot model lookup and device-flow polling client |
| `TimeoutHTTPLong` | 15 min | Overall timeout of request clients |
| `TimeoutWait` | 10 min | Upper bound of device code polling (Codex, Copilot) and the Gemini stream idle window |
| `IntervalBackoff` | 5 s | Device code poll interval when the server sends none, and the step added on Copilot `slow_down` |
| `ExpiryToken` | 1 h | Codex / Grok token lifetime when the token response has no `expires_in`, and the Gemini context-cache TTL |
| `ExpirySafety` | 60 s | Margin before expiry at which tokens refresh and Gemini caches are treated as expired |

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
