# 供應商憑證

各供應商前綴需要哪些 `router.Config` 欄位。

## router.Config

函式庫本身不讀任何環境變數，憑證一律由呼叫端傳入。

```go
type Config struct {
	Name    string // provider@model
	APIKey  string
	Token   any    // *llmrouter.CopilotToken / *llmrouter.CodexToken / *llmrouter.GrokToken
	BaseURL string // compat 端點
	AccountID string // cloudflare
	GatewayID string // cloudflare
	EnableClaude bool   // claude-code
	StateDir     string // claude-code
}
```

## 各供應商需要的欄位

| 前綴 | 必要欄位 | 備註 |
|---|---|---|
| `claude@` | `APIKey` | — |
| `claude-code@` | `EnableClaude: true` | 需要 `PATH` 中已登入的 `claude` CLI；選填 `StateDir` 保存 session |
| `openai@` | `APIKey` | 新世代模型自動改走 Responses API |
| `gemini@` | `APIKey` | 前綴達 4096 token 時自動建立 `cachedContents` |
| `grok@` | `APIKey` | — |
| `deepseek@` | `APIKey` | — |
| `mistral@` | `APIKey` | — |
| `nvidia@` | `APIKey` | — |
| `ollama-cloud@` | `APIKey` | — |
| `openrouter@` | `APIKey` | — |
| `cloudflare@` | `APIKey` + `AccountID` | 走 AI Gateway；`GatewayID` 選填，留空時預設 `default` |
| `compat@` / `compat[name]@` | `BaseURL` | `BaseURL` 留空時預設 `http://localhost:11434/v1`；`APIKey` 選填，留空時不送 `Authorization` header |
| `copilot@` | `Token`（`*llmrouter.CopilotToken`） | 另需換取短期 session token |
| `codex@` | `Token`（`*llmrouter.CodexToken`） | 帶 `ChatGPT-Account-Id` header |
| `grok-oauth@` | `Token`（`*llmrouter.GrokToken`） | — |

## 相關頁面

- [Claude Code CLI](/zh/core-concepts-claude-code)：`EnableClaude` 與 `StateDir`
- [OAuth 權杖](/zh/configuration-oauth)：取得 Copilot／Codex／Grok 的 `Token`
- [自訂端點](/zh/configuration-endpoints)：`BaseURL`、`AccountID`、`GatewayID` 範例
