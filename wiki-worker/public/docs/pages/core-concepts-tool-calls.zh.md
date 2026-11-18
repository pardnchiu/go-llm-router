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

## 伺服器端工具：X 搜尋

`x_search` 是在 xAI 端執行的工具。一次請求中，Grok 可依關鍵字或語意搜尋 X、查詢使用者、讀取完整討論串，再根據結果回答。你的程式不必執行它，也不會收到它的工具呼叫。

| 項目 | 行為 |
|---|---|
| 供應商 | `grok`（API key）與 `grok-oauth` |
| 模型 | 所有 `grok-*` 對話模型；`imagine` 圖片與影片模型除外（`xai.SupportsXSearch`） |
| 何時附加 | 請求已宣告至少一個 function 工具時才附加；沒帶工具的請求不會附加 |
| 收到什麼 | 最終回答文字與你自己的 function 呼叫；搜尋步驟以 `custom_tool_call` 項目回傳，`Send` 與 `SendStream` 都會略過 |
| 費用 | xAI 對每次搜尋另外計費，與 token 分開；一個問題可能觸發多次，實測同一題依模型搜尋 3 到 27 次 |
| 關閉方式 | 目前沒有選項；請求不帶工具，或改用 Grok 以外的供應商 |

```go
agent, _ := router.New(router.Config{
	Name:   "grok@grok-4.3",
	APIKey: os.Getenv("GROK_API_KEY"),
})
out, _, _ := agent.Send(ctx, []llmrouter.Message{
	{Role: "user", Content: "What did @xai post this week?"},
}, tools, llmrouter.ReasoningDefault, llmrouter.ModeDefault)
```

`tools` 為任意非空的 function 工具清單，例如上方的 `get_weather`；有帶工具才會附加 `x_search`。

## 相關頁面

- [訊息型別](/zh/api-reference-messages)：`Tool`、`ToolCall`、`ToolCallFunction` 欄位
- [串流](/zh/core-concepts-streaming)：工具呼叫以 `tool_call` 事件抵達
