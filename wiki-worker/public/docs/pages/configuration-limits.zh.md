# 逾時與上限

函式庫套用的請求逾時與回應讀取上限。

## 數值

| 項目 | 值 | 來源 |
|---|---|---|
| 請求 timeout | 10 分鐘 | `llmrouter.NewHTTPClient` |
| Response header timeout | 15 秒 | `codex`、`grok-oauth` 自建 client，以及 Codex／Grok OAuth 刷新用 client |
| 模型清單／用量查詢 timeout | 10 秒 | 各走 HTTP 的 `Models` / `Usage` 內建；`claudeCode` 執行 CLI，本身不設 timeout |
| 串流停滯／重複輸出監控 | 連續 3 次探測失敗，首次 30 秒、之後每 10 秒（共 50 秒） | `llmrouter.WatchIdle`，見[串流](/zh/core-concepts-streaming#串流監控) |
| 非串流音訊 body 上限 | 64 MiB；SSE 串流不設上限 | `llmrouter.StreamBodyLimit` |
| 錯誤 body 上限 | 8 KiB | `llmrouter.ErrorBodyLimit` |

## 相關頁面

- [串流符號](/zh/api-reference-streaming)：所有讀取上限常數
