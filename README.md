擷取自 `Agenvoy` [2741d4a](https://github.com/agenvoy/Agenvoy/commit/2741d4a3be70c5bfac1bba9e1d5d54a65db15acd)

> **v0.28.17 起**，Agenvoy 內的 LLM provider 統一改用本套件 `go-llm-router`，不再各自維護 provider 邏輯。

## 安裝

```bash
go get github.com/pardnchiu/go-llm-router
```

## 套件結構

```
github.com/pardnchiu/go-llm-router/core           // provider.Agent 介面、共用型別（Message／Tool／Output／Usage）
github.com/pardnchiu/go-llm-router/core/router     // 統一入口：依 provider 名稱建立對應 Agent
github.com/pardnchiu/go-llm-router/core/claude     // Anthropic Claude
github.com/pardnchiu/go-llm-router/core/openai     // OpenAI
github.com/pardnchiu/go-llm-router/core/openaiCodex // OpenAI Codex（OAuth）
github.com/pardnchiu/go-llm-router/core/gemini     // Google Gemini
github.com/pardnchiu/go-llm-router/core/grok       // xAI Grok（API Key）
github.com/pardnchiu/go-llm-router/core/grokOauth  // xAI Grok（OAuth）
github.com/pardnchiu/go-llm-router/core/deepseek   // DeepSeek
github.com/pardnchiu/go-llm-router/core/nvidia     // NVIDIA NIM
github.com/pardnchiu/go-llm-router/core/openRouter // OpenRouter
github.com/pardnchiu/go-llm-router/core/cloudflare // Cloudflare Workers AI
github.com/pardnchiu/go-llm-router/core/copilot    // GitHub Copilot（OAuth）
github.com/pardnchiu/go-llm-router/core/compat     // 自訂 OpenAI-compatible endpoint
github.com/pardnchiu/go-llm-router/core/oauth      // 各 provider 的 OAuth 登入／token refresh 流程
```

## 基本用法

透過 `core/router` 依 provider 名稱字串建立對應的 `provider.Agent`：

```go
package main

import (
	"context"
	"fmt"

	provider "github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/router"
)

func main() {
	agent, err := router.New(router.Config{
		Name:   "claude@claude-sonnet-5", // provider@model
		APIKey: "sk-ant-...",
	})
	if err != nil {
		panic(err)
	}

	messages := []provider.Message{
		{Role: "user", Content: "你好"},
	}

	output, statusCode, err := agent.Send(context.Background(), messages, nil, "medium")
	if err != nil {
		panic(err)
	}

	fmt.Println(statusCode, output.Choices[0].Message.Content)
}
```

### `Config.Name` 格式

`router.New` 以 `config.Name` 判斷 provider 與 model：

| 格式 | 說明 | 範例 |
|---|---|---|
| `<provider>@<model>` | 標準格式，`@` 前為 provider key、後為 model 名稱 | `openai@gpt-5.4` |
| `<provider>[...]@<model>` | provider 名稱後可帶中括號附加資訊（會被忽略，僅用於路由比對） | `claude[eu]@claude-opus-4-8` |
| `compat@<model>` | `compat` provider 專用，需另外指定 `BaseURL` | `compat@my-model` |

支援的 `provider` key：`claude`、`openai`、`gemini`、`grok`、`grok-oauth`、`deepseek`、`nvidia`、`openrouter`、`cloudflare`、`compat`、`copilot`、`codex`。

### `router.Config` 欄位

| 欄位 | 用途 |
|---|---|
| `Name` | `<provider>@<model>` |
| `APIKey` | API Key（`claude`／`openai`／`gemini`／`grok`／`deepseek`／`nvidia`／`openrouter`／`cloudflare`／`compat`） |
| `Token` | OAuth token（`copilot`／`codex`／`grok-oauth`），型別依 provider 為 `*provider.CopilotToken`／`*provider.CodexToken`／`*provider.GrokToken` |
| `BaseURL` | 僅 `compat` 使用，自訂 OpenAI-compatible endpoint |
| `AccountID` / `GatewayID` | 僅 `cloudflare` 使用 |

## 直接使用單一 provider

不透過 `router`，也可直接呼叫個別 provider 套件：

```go
import (
	provider "github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/claude"
)

agent, err := claude.New(provider.Config{
	Model:  "claude-sonnet-5",
	APIKey: "sk-ant-...",
})
```

## OAuth Provider（Copilot／Codex／Grok）

`copilot`、`codex`、`grok-oauth` 不使用 API Key，改用 `core/oauth` 下對應套件完成登入並取得 `Token`，再傳入 `router.Config.Token` 或各 provider 的 `provider.Config.Token`。Token 內建 `Expired()` 判斷是否需要 refresh。

## `provider.Agent` 介面

所有 provider 皆實作同一介面：

```go
type Agent interface {
	Name() string
	Send(ctx context.Context, messages []Message, toolDefs []Tool, reasoning string) (*Output, int, error)
}
```

- `reasoning`：`none` / `low` / `medium` / `high` / `xhigh`，各 provider 是否支援與如何對應由 `core/provider.go`（`SupportsReasoningSwitch`、`ClampReasoningLevel` 等）統一處理

## Usage（token 用量）

`Output.Usage` 透過自訂 `UnmarshalJSON` 自動吸收各家 provider 不同的欄位命名（`input_tokens`／`prompt_tokens`、`output_tokens`／`completion_tokens`、`prompt_tokens_details.cached_tokens` 等），統一正規化為以下欄位：

| 欄位 | 說明 |
|---|---|
| `Input` | 輸入 token 數（已扣除 cache 命中部分） |
| `Output` | 輸出 token 數 |
| `CacheCreate` | 建立 prompt cache 所消耗的 token 數（`cache_creation_input_tokens`） |
| `CacheRead` | 命中 prompt cache 的 token 數（`cache_read_input_tokens` 或 `prompt_tokens_details.cached_tokens`） |

```go
output, _, err := agent.Send(ctx, messages, nil, "medium")
if err != nil {
	panic(err)
}

usage := output.Usage
fmt.Printf("input=%d output=%d cache_create=%d cache_read=%d\n",
	usage.Input, usage.Output, usage.CacheCreate, usage.CacheRead)
```

不需要依 provider 各自解析回應格式，呼叫端只需讀取正規化後的 `output.Usage` 即可。
