# Timeouts and Limits

The request timeouts and response read caps the library applies.

## Values

| Item | Value | Source |
|---|---|---|
| Request timeout | 15 minutes | `llmrouter.TimeoutHTTPLong`, via `NewHTTPClient` / `NewStreamHTTPClient` |
| Response-header timeout | 15 seconds | `llmrouter.TimeoutHTTPShort`, via `NewStreamHTTPClient` (`codex`, `grok-oauth`, Codex / Grok OAuth requests) |
| Model listing / usage query timeout | 15 seconds | `llmrouter.TimeoutHTTPShort` inside each HTTP-based `Models` / `Usage`; `claudeCode` runs the CLI with no timeout of its own |
| Stream idle / repetition watchdog | 3 failed probes, first at 30 s then every 10 s (50 s total); Gemini first probes at 10 minutes | `llmrouter.WatchIdle` / `SetIdleWindow`, see [Streaming](/core-concepts-streaming#stream-watchdog) |
| Device code login wait | Up to 10 minutes, polling every 5 s unless the server sets an interval | `llmrouter.TimeoutWait` / `IntervalBackoff`, see [OAuth Tokens](/configuration-oauth#logging-in-without-a-local-browser) |
| Gemini voice summary | 5 minutes | `summary.VoiceReply` |
| Non-streamed audio body cap | 64 MiB; SSE streams are not capped | `llmrouter.StreamBodyLimit` |
| Error body cap | 8 KiB | `llmrouter.ErrorBodyLimit` |

## Related

- [Stream Symbols](/api-reference-streaming): every read-cap constant
