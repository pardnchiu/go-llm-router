# Timeouts and Limits

The request timeouts and response read caps the library applies.

## Values

| Item | Value | Source |
|---|---|---|
| Request timeout | 10 minutes | `llmrouter.NewHTTPClient` |
| Response-header timeout | 15 seconds | the clients built by `codex` and `grok-oauth`, and the Codex / Grok OAuth refresh client |
| Model listing / usage query timeout | 10 seconds | inside each HTTP-based `Models` / `Usage`; `claudeCode` runs the CLI with no timeout of its own |
| Stream idle / repetition watchdog | 3 failed probes, first at 30 s then every 10 s (50 s total) | `llmrouter.WatchIdle`, see [Streaming](/core-concepts-streaming#stream-watchdog) |
| Non-streamed audio body cap | 64 MiB; SSE streams are not capped | `llmrouter.StreamBodyLimit` |
| Error body cap | 8 KiB | `llmrouter.ErrorBodyLimit` |

## Related

- [Stream Symbols](/api-reference-streaming): every read-cap constant
