# Provider Credentials

Which `router.Config` fields each provider prefix needs.

## router.Config

The library reads no environment variables; callers pass every credential in.

```go
type Config struct {
	Name    string // provider@model
	APIKey  string
	Token   any    // *llmrouter.CopilotToken / *llmrouter.CodexToken / *llmrouter.GrokToken
	BaseURL string // compat endpoint
	AccountID string // cloudflare
	GatewayID string // cloudflare
	EnableClaude bool   // claude-code
	StateDir     string // claude-code
}
```

## Required fields per provider

| Prefix | Required | Notes |
|---|---|---|
| `claude@` | `APIKey` | — |
| `claude-code@` | `EnableClaude: true` | Needs a logged-in `claude` CLI on `PATH`; optional `StateDir` persists sessions |
| `openai@` | `APIKey` | Newer generations switch to the Responses API automatically |
| `gemini@` | `APIKey` | Creates `cachedContents` once a prefix reaches 4096 tokens |
| `grok@` | `APIKey` | — |
| `deepseek@` | `APIKey` | — |
| `mistral@` | `APIKey` | — |
| `nvidia@` | `APIKey` | — |
| `ollama-cloud@` | `APIKey` | — |
| `openrouter@` | `APIKey` | — |
| `cloudflare@` | `APIKey` + `AccountID` | Runs through AI Gateway; optional `GatewayID` defaults to `default` |
| `compat@` / `compat[name]@` | `BaseURL` | An empty `BaseURL` defaults to `http://localhost:11434/v1`; `APIKey` is optional and, when empty, no `Authorization` header is sent |
| `copilot@` | `Token` (`*llmrouter.CopilotToken`) | Additionally exchanges a short-lived session token |
| `codex@` | `Token` (`*llmrouter.CodexToken`) | Sends the `ChatGPT-Account-Id` header |
| `grok-oauth@` | `Token` (`*llmrouter.GrokToken`) | — |

## Related

- [Claude Code CLI](/core-concepts-claude-code): `EnableClaude` and `StateDir`
- [OAuth Tokens](/configuration-oauth): obtaining `Token` for Copilot, Codex, and Grok
- [Custom Endpoints](/configuration-endpoints): `BaseURL`, `AccountID`, `GatewayID` examples
