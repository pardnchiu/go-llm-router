# 自訂端點

如何把 router 指向自架或第三方 OpenAI 相容端點，或 Cloudflare Workers AI。

## 自架與第三方相容端點

```go
// 明確指定實例名，Name() 會回 lmstudio@qwen3-30b
agent, err := router.New(router.Config{
	Name:    "compat[lmstudio]@qwen3-30b",
	BaseURL: "http://127.0.0.1:1234/v1",
	APIKey:  "not-needed",
})

// 未知前綴自動落到 compat，Name() 保留原前綴
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

模型清單以 Workers AI 的 task 名稱篩選，而非模型 ID 標記，因此 `ModelFilter{STTOnly: true}` 會取 `Automatic Speech Recognition` 這個 task。

## 相關頁面

- [模型路由](/zh/core-concepts-routing)：未知前綴如何被改寫
