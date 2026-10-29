# 工具呼叫

如何在所有供應商之間宣告工具並回傳工具結果。

## 宣告工具

工具定義採 OpenAI 形狀，`Parameters` 是 `json.RawMessage`，直接放 JSON Schema：

```go
tools := []llmrouter.Tool{{
	Type: "function",
	Function: llmrouter.ToolFunction{
		Name:        "get_weather",
		Description: "Get the current weather for a city",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}`),
	},
}}
```

## 回傳結果

回傳結果以 `Role: "tool"` 搭配 `ToolCallID` 帶回，`ToolCallID` 即 assistant 回合中 `ToolCall.ID` 的原值。Gemini 本身沒有呼叫 ID：轉接層把 `ToolCall.ID` 設為函式名稱，並將 `ToolCallID` 作為 `functionResponse` 的 name 送回。Gemini 另外會在 `ToolCall.ThoughtSignature` 附上簽章，**必須原樣回傳**，否則下一回合會被拒絕。

## 相關頁面

- [訊息型別](/zh/api-reference-messages)：`Tool`、`ToolCall`、`ToolCallFunction` 欄位
- [串流](/zh/core-concepts-streaming)：工具呼叫以 `tool_call` 事件抵達
