# 輔助函式

Agent 介面之外的獨立輔助函式。

## HTTP 與模型判斷

| 符號 | 簽名 | 說明 |
|---|---|---|
| `NewHTTPClient` | `func() *http.Client` | 10 分鐘 timeout |
| `SupportTemperature` | `func(providerName, model string) bool` | Claude、`gpt-5` 系列、`deepseek-reasoner`、Gemini preview 不接受 |
| `ResponsesAPI` | `func(providerName, model string) bool` | 決定走 Responses 或 Chat Completions |
| `summary.VoiceReply`（`core/gemini/summary`） | `func(ctx context.Context, text string) (string, error)` | 超過 320 字元時以 Gemini 壓成可朗讀版本 |

## Session

| 符號 | 簽名 | 說明 |
|---|---|---|
| `WithSessionID` | `func(ctx context.Context, id string) context.Context` | 附加去除空白後的 session ID；ID 為空時原樣回傳 `ctx` |
| `SessionID` | `func(ctx context.Context) string` | 已附加的 ID，沒有則回 `""` |
| `SessionUUID` | `func(ctx context.Context) string` | 以 SHA-256 雜湊 ID 並格式化為 UUID；未附加 ID 時回 `""` |
| `WithCacheTTL` | `func(ctx context.Context, ttl string) context.Context` | 附加去除空白後的 prompt 快取 TTL（如 `"1h"`）；值為空時原樣回傳 `ctx` |
| `CacheTTL` | `func(ctx context.Context) string` | 已附加的 TTL，沒有則回 `""` |

## 相關頁面

- [對話快取鍵](/zh/core-concepts-sessions)：哪些供應商使用此鍵與 TTL
