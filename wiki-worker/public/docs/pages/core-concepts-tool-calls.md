# Tool Calls

How to declare tools and return tool results across every provider.

## Declaring tools

Tool definitions use the OpenAI shape, with `Parameters` as `json.RawMessage` so a JSON Schema goes in verbatim:

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

## Returning results

Results come back as `Role: "tool"` plus `ToolCallID`, where `ToolCallID` is the `ToolCall.ID` from the assistant turn, passed back unchanged. Gemini has no call IDs of its own: the adapter sets `ToolCall.ID` to the function name and sends `ToolCallID` back as the `functionResponse` name. Gemini additionally attaches `ToolCall.ThoughtSignature`, which **must be echoed back verbatim** or the next turn is rejected.

## Server-side tools: X search

`x_search` is a tool that runs on xAI's side. During one request, Grok can search X by keyword or meaning, look up users, and read whole threads, then answer with what it found. Your code never executes it and never sees a tool call for it.

| Item | Behavior |
|---|---|
| Providers | `grok` (API key) and `grok-oauth` |
| Models | Every `grok-*` chat model; `imagine` image and video models are excluded (`xai.SupportsXSearch`) |
| When it is attached | Only when the request already declares at least one function tool; a request without tools sends none |
| What you receive | The final answer text and your own function calls; search steps arrive as `custom_tool_call` items, which both `Send` and `SendStream` drop |
| Cost | xAI bills each search separately from tokens, and one question can trigger many: a single test question made 3 to 27 searches depending on the model |
| Turning it off | No option yet; send the request without tools, or use a provider other than Grok |

```go
agent, _ := router.New(router.Config{
	Name:   "grok@grok-4.3",
	APIKey: os.Getenv("GROK_API_KEY"),
})
out, _, _ := agent.Send(ctx, []llmrouter.Message{
	{Role: "user", Content: "What did @xai post this week?"},
}, tools, llmrouter.ReasoningDefault, llmrouter.ModeDefault)
```

`tools` is any non-empty function tool list, such as the `get_weather` example above; its presence is what attaches `x_search`.

## Related

- [Message Types](/api-reference-messages): `Tool`, `ToolCall`, `ToolCallFunction` fields
- [Streaming](/core-concepts-streaming): tool calls arriving as `tool_call` events
