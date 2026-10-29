# Fast Mode

How `ModeFast` maps onto each vendor's priority tier and how a silent downgrade is surfaced.

## Checking support

Support depends on the model generation, so check before sending:

```go
mode := llmrouter.ModeDefault
if llmrouter.SupportFast("claude", "claude-opus-5") {
	mode = llmrouter.ModeFast
}
```

## Vendor mapping

An adapter applies `ModeFast` only when `SupportFast` is true for its model; otherwise the request goes out at the default tier.

| Provider | Request change | `SupportFast` is true for |
|---|---|---|
| `claude` | `"speed": "fast"` plus `anthropic-beta: fast-mode-2026-02-01` | models containing `opus-5` or `opus-4-8` |
| `openai` | `service_tier: "priority"` | GPT 5.4+ (codex models from 5.3), excluding `-pro` and `-nano` |
| `grok`, `grok-oauth` | `service_tier: "priority"` | every model except `imagine` ones |
| `openrouter` | `service_tier: "priority"` | `openai/`, `google/`, `x-ai/` models passing the matching vendor rule |

`SupportFast("gemini", ...)` returns true for the `gemini-3.5-flash`, `gemini-3.1-flash-lite`, `gemini-3.1-pro`, `gemini-3-flash`, `gemini-2.5-pro`, and `gemini-2.5-flash` prefixes, but only `openrouter` acts on it; the `gemini` adapter ignores `Mode`, as does every other provider not in the table.

## Downgrade warning

After a fast request, the adapter passes the served tier (`service_tier`, or Claude's `usage.speed`) to `WarnFastDowngrade`; when it is not empty, `fast`, or `priority`, a warning is logged — a request billed as fast but downgraded silently would leave both cost and latency unaccounted for.

## Related

- [Reasoning Levels](/core-concepts-reasoning)
- [Reasoning and Modes](/api-reference-reasoning): signatures
