# 對話快取鍵

附在 context 上的 session ID 如何轉成各供應商的 prompt 快取鍵或 session affinity header。

## 附加 session

```go
ctx = llmrouter.WithSessionID(ctx, conversationID)
out, code, err := agent.Send(ctx, messages, nil, llmrouter.ReasoningDefault, llmrouter.ModeDefault)
```

`WithSessionID` 會先去除前後空白，ID 為空時原樣回傳 context，因此傳入未設定的 ID 也安全。`SessionID(ctx)` 讀回原值，context 為 nil 時回 `""`。

## 衍生鍵

轉接層不會送出原始 ID。`SessionUUID(ctx)` 以 SHA-256 雜湊後，把前 128 bits 格式化成 UUID 形狀的字串（version 位元為 `8`、variant 位元為 `a`）。同一段對話每次請求都得到相同的鍵，而呼叫端自己的識別碼不會離開程序。

## 各供應商對映

| 供應商 | 鍵放在哪裡 | 適用範圍 |
|---|---|---|
| `openai` | body `prompt_cache_key` | Chat Completions 與 Responses，`Send` 與 `SendStream` |
| `codex` | header `session_id` 與 body `prompt_cache_key` | `Send` 與 `SendStream` |
| `grok`、`grok-oauth` | body `prompt_cache_key`（由 `core/xai` 組出） | `Send` 與 `SendStream` |
| `cloudflare` | header `x-session-affinity` | `Send`（不支援串流） |
| `claude-code` | 不送出；用於選出該對話的常駐 CLI 程序 | `Send`（見 [Claude Code CLI](/zh/core-concepts-claude-code)） |
| 其他供應商 | 忽略 | — |

## 快取 TTL

```go
ctx = llmrouter.WithCacheTTL(ctx, "1h")
```

`WithCacheTTL` 以相同方式附加 prompt 快取 TTL：先去除前後空白，值為空時原樣回傳 context，`CacheTTL(ctx)` 讀回原值。目前只有 `claude-code` 讀取此值：`1h` 啟用 CLI 的一小時 prompt 快取，並在設定 `StateDir` 時保存 session；其他值使用預設的五分鐘。

## 相關頁面

- [用量正規化](/zh/core-concepts-usage)：快取命中會計入 `CacheRead`
- [輔助函式](/zh/api-reference-helpers#session)：函式簽名
