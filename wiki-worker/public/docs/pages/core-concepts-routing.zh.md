# 模型路由

`router.New` 如何把 `provider@model` 名稱轉成 Agent，以及如何換供應商。

## provider@model 解析

`router.New` 以 `@` 切開名稱，前綴查表決定建構哪個 Agent：

| 輸入 | 結果 |
|---|---|
| `claude@claude-opus-5` | `core/claude` 的 Agent |
| `claude-code@sonnet` | `core/claudeCode` 的 Agent，需 `EnableClaude: true` |
| `compat[lmstudio]@qwen3-30b` | `core/compat`，`Name()` 回 `lmstudio@qwen3-30b` |
| `groq@llama-3.3-70b`（未知前綴） | 改寫為 `compat[groq]@llama-3.3-70b`，`Name()` 回 `groq@llama-3.3-70b` |
| `groq`（無 `@`） | 回 `unknown provider` 錯誤 |

已知前綴即 `core` 中的 `PROVIDER_*` 常數；`llmrouter.Providers()` 列出除 `compat` 以外的全部前綴。方括號實例名讓同一個 compat 轉接層同時代表多個端點，`Name()` 會保留實例名，log 與用量統計才分得出來源。

## 換一家供應商

`Name` 是唯一要改的地方，其餘程式碼不動：

```go
agent, err := router.New(router.Config{Name: "openai@gpt-5.1", APIKey: os.Getenv("OPENAI_API_KEY")})
agent, err = router.New(router.Config{Name: "gemini@gemini-3-pro", APIKey: os.Getenv("GEMINI_API_KEY")})
agent, err = router.New(router.Config{Name: "deepseek@deepseek-chat", APIKey: os.Getenv("DEEPSEEK_API_KEY")})
```

## 未知前綴

未知前綴會自動落到 OpenAI 相容轉接層。`BaseURL` 設為該廠商端點；未設定時打向 `http://localhost:11434/v1`：

```go
agent, err := router.New(router.Config{
	Name:    "groq@llama-3.3-70b",
	BaseURL: "https://api.groq.com/openai/v1",
	APIKey:  os.Getenv("GROQ_API_KEY"),
})
```

## 相關頁面

- [供應商憑證](/zh/configuration)：各前綴需要的欄位
- [自訂端點](/zh/configuration-endpoints)：compat 與 Cloudflare 設定
