# fast 模式

`ModeFast` 如何對映各家的優先處理層級，以及靜默降級如何被揭露。

## 判斷是否支援

是否支援由模型世代決定，`SupportFast` 先判斷再送：

```go
mode := llmrouter.ModeDefault
if llmrouter.SupportFast("claude", "claude-opus-5") {
	mode = llmrouter.ModeFast
}
```

## 各家對映

轉接層只在 `SupportFast` 對該模型回傳 true 時套用 `ModeFast`，否則以預設層級送出。

| 供應商 | 請求變更 | `SupportFast` 為 true 的模型 |
|---|---|---|
| `claude` | `"speed": "fast"` 加上 `anthropic-beta: fast-mode-2026-02-01` | 名稱含 `opus-5` 或 `opus-4-8` |
| `openai` | `service_tier: "priority"` | GPT 5.4 以上（codex 模型自 5.3 起），排除 `-pro` 與 `-nano` |
| `grok`、`grok-oauth` | `service_tier: "priority"` | `imagine` 以外的全部模型 |
| `openrouter` | `service_tier: "priority"` | 通過對應廠商規則的 `openai/`、`google/`、`x-ai/` 模型 |

`SupportFast("gemini", ...)` 對 `gemini-3.5-flash`、`gemini-3.1-flash-lite`、`gemini-3.1-pro`、`gemini-3-flash`、`gemini-2.5-pro`、`gemini-2.5-flash` 前綴回傳 true，但只有 `openrouter` 會據此動作；`gemini` 轉接層與表外其他供應商都忽略 `Mode`。

## 降級警告

fast 請求完成後，轉接層把實際層級（`service_tier`，Claude 則為 `usage.speed`）交給 `WarnFastDowngrade`；值不是空字串、`fast` 或 `priority` 時發出警告 —— 付了 fast 的請求卻被降級，靜默吞掉會讓成本與延遲都對不上帳。

## 相關頁面

- [推理等級](/zh/core-concepts-reasoning)
- [推理與模式符號](/zh/api-reference-reasoning)：簽名
