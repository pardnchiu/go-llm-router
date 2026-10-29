# 推理等級

`llmrouter.Reasoning` 六個等級如何對映到各家，以及模型無法支援時的處理。

## 等級

各等級對映到各家的不同機制：Claude 的 thinking budget（4.5／4.1）或 adaptive thinking effort（較新模型）、Gemini 的 thinking config、OpenAI 的 reasoning effort。

| 等級 | 字串 | 別名 |
|---|---|---|
| `ReasoningNone` | `none` | — |
| `ReasoningLow` | `low` | `minimal` |
| `ReasoningMedium` | `medium`（預設） | — |
| `ReasoningHigh` | `high` | — |
| `ReasoningXHigh` | `xhigh` | `extra` |
| `ReasoningMax` | `max` | `ultra` |

## 收斂到模型範圍

超出模型支援範圍的請求不會被拒絕，而是收斂到最接近的合法值並輸出 debug log。實作 `ReasoningAgent` 的轉接層都已在 `Send` 內部收斂；自行呼叫 `ClampReasoning` 只是為了事先得知實際等級：

```go
if limited, ok := agent.(llmrouter.ReasoningAgent); ok {
	low, high := limited.ReasoningLimits()
	reasoning = llmrouter.ClampReasoning(reasoning, low, high, "openai", "gpt-5.1")
}
```

`deepseek`、`cloudflare`、`compat` 忽略此等級；`claude-code` 以 effort 傳給 CLI，`none` 會以 `low` 送出。

收斂而非報錯的理由：呼叫端通常是一個對所有模型共用的設定值，讓上限比較低的模型直接跑完，比讓整個請求失敗更符合實際使用。

## 相關頁面

- [fast 模式](/zh/core-concepts-fast-mode)：另一個逐請求設定
- [推理與模式符號](/zh/api-reference-reasoning)：簽名
