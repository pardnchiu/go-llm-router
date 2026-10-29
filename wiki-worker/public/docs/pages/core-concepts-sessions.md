# Session Cache Keys

How a session ID attached to the context becomes a per-provider prompt-cache key or session-affinity header.

## Attaching a session

```go
ctx = llmrouter.WithSessionID(ctx, conversationID)
out, code, err := agent.Send(ctx, messages, nil, llmrouter.ReasoningDefault, llmrouter.ModeDefault)
```

`WithSessionID` trims the ID and returns the context unchanged when it is empty, so passing an unset ID is safe. `SessionID(ctx)` reads it back and returns `""` for a nil context.

## Derived key

Adapters never send the raw ID. `SessionUUID(ctx)` hashes it with SHA-256 and formats the first 128 bits as a UUID-shaped string (version nibble `8`, variant nibble `a`). The same conversation therefore maps to the same key on every request, while the caller's own identifier never leaves the process.

## Per-provider mapping

| Provider | Where the key goes | Applies to |
|---|---|---|
| `openai` | body `prompt_cache_key` | Chat Completions and Responses, `Send` and `SendStream` |
| `codex` | header `session_id` and body `prompt_cache_key` | `Send` and `SendStream` |
| `grok`, `grok-oauth` | body `prompt_cache_key` (built by `core/xai`) | `Send` and `SendStream` |
| `cloudflare` | header `x-session-affinity` | `Send` (streaming is unsupported) |
| `claude-code` | not sent; selects the long-lived CLI process for the conversation | `Send` (see [Claude Code CLI](/core-concepts-claude-code)) |
| every other provider | ignored | — |

## Cache TTL

```go
ctx = llmrouter.WithCacheTTL(ctx, "1h")
```

`WithCacheTTL` attaches a prompt-cache TTL the same way: it trims the value and returns the context unchanged when it is empty, and `CacheTTL(ctx)` reads it back. Only `claude-code` reads it: `1h` sets the CLI's one-hour prompt cache and, with a `StateDir`, persists the session; any other value uses the five-minute default.

## Related

- [Usage Normalization](/core-concepts-usage): cache hits show up as `CacheRead`
- [Helpers](/api-reference-helpers#session): function signatures
