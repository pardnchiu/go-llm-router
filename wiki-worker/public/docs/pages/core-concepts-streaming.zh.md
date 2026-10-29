# 串流

如何以正規化事件串流接收回應，以及串流失敗如何回報。

## 串流輸出

串流是選用能力，以型別斷言取得：

```go
streamAgent, ok := agent.(llmrouter.StreamAgent)
if !ok {
	log.Fatalf("%s 不支援串流", agent.Name())
}

events, err := streamAgent.SendStream(ctx, messages, nil, llmrouter.ReasoningHigh, llmrouter.ModeDefault)
if err != nil {
	log.Fatal(err)
}

for evt := range events {
	switch evt.Type {
	case llmrouter.StreamEventText:
		fmt.Print(evt.TextDelta)
	case llmrouter.StreamEventUsage:
		fmt.Printf("\n[usage] in=%d out=%d\n", evt.Usage.Input, evt.Usage.Output)
	case llmrouter.StreamEventError:
		log.Fatal(evt.Err)
	}
}
```

## 事件型別

Chat Completions 與 Responses 兩種上游 SSE 都會被正規化成同一組事件，呼叫端不需要分辨來源。

```go
const (
	StreamEventText      StreamEventType = "text"
	StreamEventReasoning StreamEventType = "reasoning"
	StreamEventToolCall  StreamEventType = "tool_call"
	StreamEventUsage     StreamEventType = "usage"
	StreamEventDone      StreamEventType = "done"
	StreamEventError     StreamEventType = "error"
)
```

## 工具呼叫增量

`tool_call` 事件攜帶以 `Index` 區分的 `ToolCallDelta`。同一 index 的第一個事件帶 `ID` 與 `Name`，之後的事件只帶 `Arguments` 片段，因此依 `Index` 串接 `Arguments` 即可。Gemini 每個呼叫以單一事件完整送出，並帶上 `ThoughtSignature`。Responses 串流中，若參數只出現在最後的 `response.output_item.done` 項目而沒有以增量送出，會在該處補發一次，依 index 串接仍可得到完整 JSON。`claude` 請求會在每個工具設定 `eager_input_streaming`，大型工具參數（例如整份 HTML）會以連續增量送達，而不是在伺服器緩衝完後一次送出。

## 串流錯誤

開啟串流失敗時，`SendStream` 回傳 `*llmrouter.StreamError`（帶 `Provider` / `Code` / `Body`）。非 200 狀態會填入 `Code` 與 `Body`；200 但 `Content-Type` 不是 SSE 時，另以 `llmrouter.ErrStreamUnsupported` 作為 `Unwrap` 目標，因此 `errors.As` 取得狀態碼、`errors.Is` 判斷能力缺失可以並存。`cloudflare` 沒有串流能力，直接回傳包裝 `ErrStreamUnsupported` 的錯誤。串流開啟後的失敗（監控逾時、無法解析的 frame、上游錯誤 frame、讀取錯誤）會以一個 `StreamEventError` 事件送出，之後 channel 關閉。串流 body 不設大小上限；錯誤訊息中的 frame 截至 512 bytes。

## 串流監控

每個 SSE 回應在串流期間都受監控。串流開啟 30 秒後進行第一次探測，之後每 10 秒一次；連續三次探測失敗即結束串流：

| 條件 | 錯誤 |
|---|---|
| 沒有收到任何 bytes | `[Timeout] <label> stream: no data for 50s` |
| 收到 frame 但都沒有文字、推理或工具呼叫內容（例如只有 `ping` 事件） | `[Timeout] <label> stream: data without parsable content for 50s` |
| 最新 16 KiB 的文字、推理與工具參數只用到 4 種以下的 byte，或在重複 4 KiB 以內的片段（每新增 512 bytes 檢查一次） | `[Timeout] <label> stream: degenerate repetition in last 16384 bytes: "<最後 128 bytes>"` |

三種錯誤都包裝 `context.DeadlineExceeded`，呼叫端可用 `errors.Is(err, context.DeadlineExceeded)` 比照一般逾時重試。只有 parser 正卡在讀取 body 時探測才算失敗，因此等待慢速 consumer 的時間不會觸發；SSE keep-alive comment（以 `:` 開頭的行）也會維持串流存活。

無法解析的 frame 會被略過，連續三個才以 `[Upstream] <label> stream decode: 3 consecutive frames` 結束串流。

## 相關頁面

- [串流符號](/zh/api-reference-streaming)：`StreamEvent` 欄位與 SSE 輔助函式
- [逾時與上限](/zh/configuration-limits)：讀取上限
