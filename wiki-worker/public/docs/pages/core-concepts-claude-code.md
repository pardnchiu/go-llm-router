# Claude Code CLI

How the `claude-code@` provider drives the locally installed Claude Code CLI as a model backend, and how it keeps one CLI process per conversation.

## Requirements

The `claude` binary must be on `PATH` and already logged in (run `claude` once). The provider is opt-in: `router.New` fails unless `EnableClaude` is `true`.

```go
agent, err := router.New(router.Config{
	Name:         "claude-code@sonnet",
	EnableClaude: true,
	StateDir:     filepath.Join(home, ".myapp", "sessions"),
})
```

| Field | Effect |
|---|---|
| `Name` | `claude-code@<model>`; the model is passed to `claude --model` unchanged |
| `EnableClaude` | required; `false` returns `claudeCode.New: EnableClaude is required` |
| `StateDir` | optional; enables resuming a conversation after the process restarts (see below) |

`claudeCode.CheckBinary()` reports a missing CLI before anything runs, and `claudeCode.Is(name)` tells whether a `provider@model` name belongs to this provider. The agent implements only `Send`: there is no `SendStream`, `ReasoningLimits`, image, or audio support, and the `Mode` argument is ignored.

## Process per conversation

Every request runs `claude -p` in stream-JSON mode with built-in tools, MCP servers, and project settings disabled (`--tools ""`, `--strict-mcp-config`, `--safe-mode`), from the OS temp directory. Leading `system` messages are joined and sent as a `<system_prompt>` block in the first user turn.

| Context | Behavior |
|---|---|
| No session ID | A fresh process per `Send` with `--no-session-persistence`; the full history is sent and the process exits afterwards |
| `WithSessionID` set | One long-lived process per session, model, system prompt, and whether `toolDefs` is non-empty. Later calls send only the messages added since the previous turn |

The process is reused while the history the caller sends still starts with what was already sent. If the caller rewrote earlier history but kept the latest answer, only the messages after that answer are sent. A changed reasoning level or diverged history restarts the process with the full history; a changed system prompt, or switching between tools and no tools, selects a separate process, which starts with the full history. An idle process is stopped after 15 minutes.

## Prompt cache and resume

`llmrouter.WithCacheTTL(ctx, "1h")` sets `CLAUDE_CODE_PROMPT_CACHE_TTL=1h` for the process; any other value, or none, uses `5m`. Changing the TTL restarts the process.

With a session ID, a `1h` TTL, and a non-empty `StateDir`, the conversation is persisted: the process gets a Claude Code session ID, and its state is written to `<StateDir>/<sessionID>/.claude_code/<slot>.json`. After your program restarts, the next call resumes that session with `--resume` instead of resending the history. If the resumed turn fails, it starts fresh once with the full history. A state file owned by another live process is ignored, and a failed turn deletes the file.

## Reasoning and output mapping

| `Reasoning` | `--effort` |
|---|---|
| `ReasoningMedium` | `medium` |
| `ReasoningHigh` | `high` |
| `ReasoningXHigh` | `xhigh` |
| `ReasoningMax` | `max` |
| any other level | `low` |

The CLI's `result` line becomes one `Output` choice: `input_tokens`, `output_tokens`, `cache_read_input_tokens`, and `cache_creation_input_tokens` fill `Usage.Input`, `Output`, `CacheRead`, and `CacheCreate`. An `is_error` result returns its `api_error_status` as the status code; a successful turn returns status code `0`. A canceled context kills the process and returns `ctx.Err()`. Image parts with a base64 `data:image/...` URL or an `http(s)` URL are forwarded as image blocks; any other image URL becomes the text `[image omitted: unsupported image URL]`.

## Related

- [Claude Code Tool Calls](/core-concepts-claude-code-tools): how tools work without native tool support
- [Session Cache Keys](/core-concepts-sessions): `WithSessionID` and `WithCacheTTL`
- [Balance and Quota](/configuration-quota): `claudeCode.Usage` subscription limits
