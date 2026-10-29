# 串流符號

串流事件型別、串流錯誤、SSE 輔助函式與讀取上限常數。

## 符號

| 符號 | 簽名／內容 |
|---|---|
| `StreamEvent` | `Type`、`TextDelta`、`ReasoningDelta`、`ToolCall`、`Usage`、`FinishReason`、`Err` |
| `StreamEventType` | `text` / `reasoning` / `tool_call` / `usage` / `done` / `error` |
| `ToolCallDelta` | `Index`、`ID`、`Name`、`Arguments`、`ThoughtSignature` |
| `StreamError` | `Provider`、`Code`、`Body`、`Err`；實作 `Error()` 與 `Unwrap()` |
| `ErrStreamUnsupported` | `error` |
| `OpenStream` | `func(ctx, client *http.Client, url string, headers map[string]string, body map[string]any, label string) (*http.Response, error)` |
| `StreamChat` / `StreamResponses` | `func(resp *http.Response, label string) <-chan StreamEvent` |
| `ResponsesStreamError` | `func(label, data string) error` |
| `ScanSSE` | `func(reader *bufio.Reader, handle func(event, data string) bool) error` |
| `TruncateFrame` | `func(data string) string` |
| `WatchIdle` | `func(body io.ReadCloser, label string) io.ReadCloser`；為 SSE body 加上停滯與重複輸出監控；`OpenStream`、`grok`、`grok-oauth`、`codex` 會套用 |
| `TouchIdle` | `func(body io.Reader, deltas ...string)`；parser 對每個帶內容的 frame 呼叫，並把文字、推理或工具參數增量交給重複偵測；未受監控的 body 不做任何事 |
| `DecodeFrame` | `func(body io.Reader, label, data string, v any) (bool, error)`；解析單一 frame，無法解析時回 `false, nil` 略過，連續第三個回 `[Upstream]` 錯誤；未受監控的 body 第一次失敗即回錯誤 |
| 監控時序 | `StreamProbeInterval` 30 秒、`StreamRecheckInterval` 10 秒、`StreamMaxProbeFailures` 3 |
| 重複偵測 | `StreamRepeatWindow` 16 KiB、`StreamRepeatMaxPeriod` 4 KiB、`StreamRepeatMaxBytes` 4、`StreamRepeatCheckBytes` 512 |
| 上限常數 | `StreamBodyLimit` 64 MiB、`ErrorBodyLimit` 8 KiB、`ErrorFrameLimit` 512 B、`JSONBodyLimit` 64 KiB；SSE 串流不設上限，`StreamBodyLimit` 用於非串流的音訊下載 |

## 相關頁面

- [串流](/zh/core-concepts-streaming)：消費事件與處理串流錯誤
