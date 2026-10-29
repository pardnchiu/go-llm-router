# Model Routing

How `router.New` turns a `provider@model` name into an Agent, and how to switch providers.

## provider@model resolution

`router.New` splits the name on `@` and looks the prefix up in a table:

| Input | Result |
|---|---|
| `claude@claude-opus-5` | an agent from `core/claude` |
| `claude-code@sonnet` | an agent from `core/claudeCode`; needs `EnableClaude: true` |
| `compat[lmstudio]@qwen3-30b` | `core/compat`; `Name()` reports `lmstudio@qwen3-30b` |
| `groq@llama-3.3-70b` (unknown prefix) | rewritten to `compat[groq]@llama-3.3-70b`; `Name()` reports `groq@llama-3.3-70b` |
| `groq` (no `@`) | `unknown provider` error |

Known prefixes are the `PROVIDER_*` constants in `core`; `llmrouter.Providers()` lists all of them except `compat`. The bracketed instance name lets one compat adapter stand for several endpoints at once, and `Name()` keeps that instance name so logs and usage accounting can still tell them apart.

## Switch providers

`Name` is the only thing that changes:

```go
agent, err := router.New(router.Config{Name: "openai@gpt-5.1", APIKey: os.Getenv("OPENAI_API_KEY")})
agent, err = router.New(router.Config{Name: "gemini@gemini-3-pro", APIKey: os.Getenv("GEMINI_API_KEY")})
agent, err = router.New(router.Config{Name: "deepseek@deepseek-chat", APIKey: os.Getenv("DEEPSEEK_API_KEY")})
```

## Unknown prefixes

An unknown prefix falls through to the OpenAI-compatible adapter. Set `BaseURL` to the vendor's endpoint; without it the adapter targets `http://localhost:11434/v1`:

```go
agent, err := router.New(router.Config{
	Name:    "groq@llama-3.3-70b",
	BaseURL: "https://api.groq.com/openai/v1",
	APIKey:  os.Getenv("GROQ_API_KEY"),
})
```

## Related

- [Provider Credentials](/configuration): required fields per prefix
- [Custom Endpoints](/configuration-endpoints): compat and Cloudflare setup
