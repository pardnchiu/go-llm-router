# Custom Endpoints

How to point the router at a self-hosted or third-party OpenAI-compatible endpoint, or at Cloudflare Workers AI.

## Self-hosted and third-party compatible endpoints

```go
// Explicit instance name; Name() reports lmstudio@qwen3-30b
agent, err := router.New(router.Config{
	Name:    "compat[lmstudio]@qwen3-30b",
	BaseURL: "http://127.0.0.1:1234/v1",
	APIKey:  "not-needed",
})

// An unknown prefix falls through to compat and keeps its prefix in Name()
agent, err = router.New(router.Config{
	Name:    "groq@llama-3.3-70b",
	BaseURL: "https://api.groq.com/openai/v1",
	APIKey:  os.Getenv("GROQ_API_KEY"),
})
```

## Cloudflare Workers AI

```go
agent, err := router.New(router.Config{
	Name:      "cloudflare@@cf/meta/llama-3.3-70b-instruct-fp8-fast",
	APIKey:    os.Getenv("CLOUDFLARE_API_KEY"),
	AccountID: os.Getenv("CLOUDFLARE_ACCOUNT_ID"),
	GatewayID: os.Getenv("CLOUDFLARE_GATEWAY_ID"),
})
```

Model listings filter on the Workers AI task name rather than model-ID markers, so `ModelFilter{STTOnly: true}` selects the `Automatic Speech Recognition` task.

## Related

- [Model Routing](/core-concepts-routing): how unknown prefixes are rewritten
