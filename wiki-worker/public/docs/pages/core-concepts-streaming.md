# Streaming

How to receive an answer as a stream of normalized events and how stream failures are reported.

## Stream the answer

Streaming is an optional capability reached by type assertion:

```go
streamAgent, ok := agent.(llmrouter.StreamAgent)
if !ok {
	log.Fatalf("%s does not support streaming", agent.Name())
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

## Event types

Chat Completions and Responses SSE both normalize onto the same event type, so callers never tell the two upstream formats apart.

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

## Tool call deltas

`tool_call` events carry a `ToolCallDelta` keyed by `Index`. The first event for an index carries `ID` and `Name`; later events for the same index carry only `Arguments` fragments, so concatenate `Arguments` per `Index`. Gemini sends each call whole in one event, with `ThoughtSignature` set. On Responses streams, arguments that arrive only in the final `response.output_item.done` item instead of as deltas are emitted there once, so the per-index concatenation still yields the full JSON. `claude` requests set `eager_input_streaming` on every tool, so large tool arguments such as a full HTML page arrive as continuous deltas instead of one burst after the server buffers them.

## Stream errors

When opening the stream fails, `SendStream` returns a `*llmrouter.StreamError` carrying `Provider` / `Code` / `Body`. A non-200 status fills `Code` and `Body`; a 200 answer with a non-SSE `Content-Type` also unwraps to `llmrouter.ErrStreamUnsupported`, so `errors.As` for the status code and `errors.Is` for the capability gap both work. `cloudflare` has no streaming and returns an error wrapping `ErrStreamUnsupported` directly. Failures after the stream opens (watchdog timeouts, undecodable frames, upstream error frames, read errors) arrive as one `StreamEventError` event, after which the channel closes. Stream bodies are not size-capped; frames quoted in errors are truncated to 512 bytes.

## Stream watchdog

Every SSE response is watched while it streams. A probe runs 30 seconds after the stream opens and then every 10 seconds; three failed probes in a row end the stream:

| Condition | Error |
|---|---|
| No bytes arrived | `[Timeout] <label> stream: no data for 50s` |
| Frames arrived but none carried text, reasoning, or tool-call content (for example only `ping` events) | `[Timeout] <label> stream: data without parsable content for 50s` |
| The latest 16 KiB of text, reasoning, and tool arguments uses 4 or fewer distinct bytes, or repeats a segment of up to 4 KiB (checked every 512 new bytes) | `[Timeout] <label> stream: degenerate repetition in last 16384 bytes: "<last 128 bytes>"` |

All three wrap `context.DeadlineExceeded`, so `errors.Is(err, context.DeadlineExceeded)` lets a caller retry them like any other timeout. A probe only counts as failed while the parser is blocked reading the body, so time spent waiting on a slow consumer never trips it, and SSE keep-alive comments (lines starting with `:`) keep the stream alive.

Undecodable frames are skipped; three in a row end the stream with `[Upstream] <label> stream decode: 3 consecutive frames`.

## Related

- [Stream Symbols](/api-reference-streaming): `StreamEvent` fields and SSE helpers
- [Timeouts and Limits](/configuration-limits): read caps
