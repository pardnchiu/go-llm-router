# Stream Symbols

Stream event types, stream errors, SSE helpers, and read-cap constants.

## Symbols

| Symbol | Signature / contents |
|---|---|
| `StreamEvent` | `Type`, `TextDelta`, `ReasoningDelta`, `ToolCall`, `Usage`, `FinishReason`, `Err` |
| `StreamEventType` | `text` / `reasoning` / `tool_call` / `usage` / `done` / `error` |
| `ToolCallDelta` | `Index`, `ID`, `Name`, `Arguments`, `ThoughtSignature` |
| `StreamError` | `Provider`, `Code`, `Body`, `Err`; implements `Error()` and `Unwrap()` |
| `ErrStreamUnsupported` | `error` |
| `OpenStream` | `func(ctx, client *http.Client, url string, headers map[string]string, body map[string]any, label string) (*http.Response, error)` |
| `StreamChat` / `StreamResponses` | `func(resp *http.Response, label string) <-chan StreamEvent` |
| `ResponsesStreamError` | `func(label, data string) error` |
| `ScanSSE` | `func(reader *bufio.Reader, handle func(event, data string) bool) error` |
| `TruncateFrame` | `func(data string) string` |
| `WatchIdle` | `func(body io.ReadCloser, label string) io.ReadCloser`; wraps an SSE body with the idle and repetition watchdog; `OpenStream`, `grok`, `grok-oauth`, and `codex` apply it |
| `TouchIdle` | `func(body io.Reader, deltas ...string)`; a parser calls it for each frame that carried content and passes the text, reasoning, or tool-argument deltas to repetition detection; no-op on an unwatched body |
| `DecodeFrame` | `func(body io.Reader, label, data string, v any) (bool, error)`; unmarshals one frame, returns `false, nil` to skip an undecodable frame and an `[Upstream]` error on the third in a row; on an unwatched body the first failure is an error |
| Watchdog timing | `StreamProbeInterval` 30 s, `StreamRecheckInterval` 10 s, `StreamMaxProbeFailures` 3 |
| Repetition detection | `StreamRepeatWindow` 16 KiB, `StreamRepeatMaxPeriod` 4 KiB, `StreamRepeatMaxBytes` 4, `StreamRepeatCheckBytes` 512 |
| Read caps | `StreamBodyLimit` 64 MiB, `ErrorBodyLimit` 8 KiB, `ErrorFrameLimit` 512 B, `JSONBodyLimit` 64 KiB; SSE streams are not capped, `StreamBodyLimit` bounds non-streamed audio downloads |

## Related

- [Streaming](/core-concepts-streaming): consuming events and handling stream errors
