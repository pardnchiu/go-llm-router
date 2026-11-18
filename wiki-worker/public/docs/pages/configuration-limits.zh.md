# 逾時與上限

函式庫套用的請求逾時與回應讀取上限。

## 數值

| 項目 | 值 | 來源 |
|---|---|---|
| 請求 timeout | 15 分鐘 | `llmrouter.TimeoutHTTPLong`，經 `NewHTTPClient`／`NewStreamHTTPClient` |
| Response header timeout | 15 秒 | `llmrouter.TimeoutHTTPShort`，經 `NewStreamHTTPClient`（`codex`、`grok-oauth`、Codex／Grok OAuth 請求） |
| 模型清單／用量查詢 timeout | 15 秒 | 各走 HTTP 的 `Models` / `Usage` 使用 `llmrouter.TimeoutHTTPShort`；`claudeCode` 執行 CLI，本身不設 timeout |
| 串流停滯／重複輸出監控 | 連續 3 次探測失敗，首次 30 秒、之後每 10 秒（共 50 秒）；Gemini 首次探測在 10 分鐘 | `llmrouter.WatchIdle`／`SetIdleWindow`，見[串流](/zh/core-concepts-streaming#串流監控) |
| Device code 登入等待 | 最長 10 分鐘，伺服器未指定間隔時每 5 秒輪詢 | `llmrouter.TimeoutWait`／`IntervalBackoff`，見[OAuth 權杖](/zh/configuration-oauth#無本機瀏覽器時登入) |
| Gemini 語音摘要 | 5 分鐘 | `summary.VoiceReply` |
| 非串流音訊 body 上限 | 64 MiB；SSE 串流不設上限 | `llmrouter.StreamBodyLimit` |
| 錯誤 body 上限 | 8 KiB | `llmrouter.ErrorBodyLimit` |

## 相關頁面

- [串流符號](/zh/api-reference-streaming)：所有讀取上限常數
