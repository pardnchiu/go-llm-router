# 輔助函式

Agent 介面之外的獨立輔助函式。

## HTTP 與模型判斷

| 符號 | 簽名 | 說明 |
|---|---|---|
| `NewHTTPClient` | `func() *http.Client` | 整體 timeout 為 `TimeoutHTTPLong`（15 分鐘） |
| `NewStreamHTTPClient` | `func() *http.Client` | 整體 timeout 為 `TimeoutHTTPLong`，並在複製的 `http.DefaultTransport` 上設 `TimeoutHTTPShort`（15 秒）response header timeout；`codex`、`grok-oauth` 與 Codex／Grok OAuth 請求使用 |
| `SupportTemperature` | `func(providerName, model string) bool` | Claude、`gpt-5` 系列、`deepseek-reasoner`、Gemini preview 不接受 |
| `ResponsesAPI` | `func(providerName, model string) bool` | 決定走 Responses 或 Chat Completions；`copilot` 僅在模型的 `supported_endpoints` 未列 `/responses` 與 `/chat/completions` 時採用 |
| `summary.VoiceReply`（`core/gemini/summary`） | `func(ctx context.Context, text string) (string, error)` | 超過 320 字元時以 Gemini 壓成可朗讀版本 |

## 時間常數

| 符號 | 值 | 用途 |
|---|---|---|
| `TimeoutHTTPShort` | 15 秒 | Response header timeout、模型清單／用量查詢、OAuth callback server 讀寫、Copilot 模型查詢與 device flow 輪詢 client |
| `TimeoutHTTPLong` | 15 分鐘 | 請求 client 的整體 timeout |
| `TimeoutWait` | 10 分鐘 | Device code 輪詢上限（Codex、Copilot）與 Gemini 串流閒置窗口 |
| `IntervalBackoff` | 5 秒 | 伺服器未給 device code 輪詢間隔時的預設值，以及 Copilot 收到 `slow_down` 時增加的間隔 |
| `ExpiryToken` | 1 小時 | Codex／Grok 權杖回應未帶 `expires_in` 時的效期，以及 Gemini context cache TTL |
| `ExpirySafety` | 60 秒 | 權杖提前刷新、Gemini 快取提前視為過期的緩衝 |

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
