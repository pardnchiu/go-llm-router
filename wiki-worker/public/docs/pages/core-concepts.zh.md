# Agent 契約

`Agent` 是所有供應商都必須實作的唯一介面，其餘能力一律以型別斷言取得。

## Agent 介面

```go
type Agent interface {
	Name() string
	Send(ctx context.Context, messages []Message, toolDefs []Tool, reasoning Reasoning, mode Mode) (*Output, int, error)
}
```

`Send` 的第二個回傳值是上游 HTTP 狀態碼，呼叫端依狀態碼分支，不必解析錯誤字串。

## 選用能力

缺少該能力的 Agent 會斷言失敗。

| 選用介面 | 方法 | 實作者 |
|---|---|---|
| `StreamAgent` | `SendStream` | `claude-code` 以外的全部供應商；`cloudflare` 一律回傳 `ErrStreamUnsupported` |
| `ReasoningAgent` | `ReasoningLimits` | `deepseek`、`cloudflare`、`compat`、`claude-code` 以外的全部供應商 |
| `ImageAgent` | `GenerateImage` | `openai`、`codex`、`gemini`、`grok`、`grok-oauth`、`openrouter` |
| `STTAgent` | `Transcribe` | `openai`、`gemini`、`openrouter` |
| `TTSAgent` | `Speak` | `openai`、`gemini`、`openrouter` |

## 為何用介面而非旗標

能力差異由型別系統表達，呼叫端用一次斷言就知道能不能做，不需要維護一張「哪家支援什麼」的對照表。

## 相關頁面

- [模型路由](/zh/core-concepts-routing)：`router.New` 如何選轉接層
- [介面與工廠](/zh/api-reference)：完整介面簽名
