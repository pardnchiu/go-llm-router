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

## Related

- [Message Types](/api-reference-messages): `Tool`, `ToolCall`, `ToolCallFunction` fields
- [Streaming](/core-concepts-streaming): tool calls arriving as `tool_call` events
