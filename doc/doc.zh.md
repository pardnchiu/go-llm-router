# go-llm-router - 技術文件

> 返回 [README](./README.zh.md)

## 前置需求

- Go 1.25 或更高
- 至少一組供應商憑證：API Key，或 Copilot / Codex / Grok 的 OAuth token
- 若使用 `core/oauth` 的 keychain 儲存，需可寫入系統 keychain

## 安裝

### go get

```bash
go get github.com/pardnchiu/go-llm-router
```

### 原始碼

```bash
git clone https://github.com/pardnchiu/go-llm-router.git
cd go-llm-router
go build ./...
```

## 設定

### 憑證對應

| 供應商 | 憑證 | 說明 |
|--------|------|------|
| `claude`、`openai`、`gemini`、`grok`、`deepseek`、`nvidia`、`openrouter` | `APIKey` | 各家 API Key |
| `cloudflare` | `APIKey` + `AccountID` + `GatewayID` | Workers AI Gateway |
| `compat` | `APIKey` + `BaseURL` | 自訂 OpenAI-compatible endpoint |
| `copilot`、`codex`、`grok-oauth` | `Token` | 經 `core/oauth` 取得的 token 結構 |

### `router.Config.Name` 格式

| 格式 | 說明 | 範例 |
|------|------|------|
| `<provider>@<model>` | 標準路由 | `openai@gpt-5.4` |
| `<provider>[...]@<model>` | 中括號附加資訊會被忽略，僅用於路由比對 | `claude[eu]@claude-opus-4-8` |
| `compat@<model>` | 需另指定 `BaseURL` | `compat@my-model` |

支援的 provider key：`claude`、`openai`、`gemini`、`grok`、`grok-oauth`、`deepseek`、`nvidia`、`openrouter`、`cloudflare`、`compat`、`copilot`、`codex`。

## 使用方式

### 基本：透過 router 建立 Agent

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
		Name:   "claude@claude-sonnet-5",
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

### 進階：直接使用單一 provider

```go
package main

import (
	"context"
	"fmt"

	provider "github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/claude"
)

func main() {
	agent, err := claude.New(provider.Config{
		Model:  "claude-sonnet-5",
		APIKey: "sk-ant-...",
	})
	if err != nil {
		panic(err)
	}

	output, statusCode, err := agent.Send(
		context.Background(),
		[]provider.Message{{Role: "user", Content: "總結這段文字"}},
		nil,
		"high",
	)
	if err != nil {
		panic(err)
	}

	fmt.Println(statusCode, output.Choices[0].Message.Content)
	fmt.Printf("input=%d output=%d cache_create=%d cache_read=%d\n",
		output.Usage.Input, output.Usage.Output, output.Usage.CacheCreate, output.Usage.CacheRead)
}
```

### 進階：Tool Calling

```go
tools := []provider.Tool{
	{
		Type: "function",
		Function: provider.ToolFunction{
			Name:        "get_weather",
			Description: "查詢城市天氣",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}`),
		},
	},
}

output, _, err := agent.Send(ctx, messages, tools, "medium")
if err != nil {
	return err
}

for _, call := range output.Choices[0].Message.ToolCalls {
	fmt.Println(call.Function.Name, call.Function.Arguments)
}
```

### 進階：OAuth Provider

```go
import oauthCopilot "github.com/pardnchiu/go-llm-router/core/oauth/copilot"

token, err := oauthCopilot.LoginWithCallback(ctx, func(code *oauthCopilot.DeviceCode) {
	fmt.Println("請開啟", code.VerificationURI, "並輸入", code.UserCode)
})
if err != nil {
	return err
}

agent, err := router.New(router.Config{
	Name:  "copilot@gpt-5",
	Token: token,
})
```

### 進階：OpenAI-compatible endpoint

```go
agent, err := router.New(router.Config{
	Name:    "compat@my-local-model",
	APIKey:  "optional-key",
	BaseURL: "http://127.0.0.1:8080/v1",
})
```

## API 參考

### `core.Agent`

```go
type Agent interface {
	Name() string
	Send(ctx context.Context, messages []Message, toolDefs []Tool, reasoning string) (*Output, int, error)
}
```

| 方法 | 說明 |
|------|------|
| `Name` | 回傳目前模型名稱 |
| `Send` | 送出訊息與可選 tools；第三個回傳值為 HTTP status code |

`reasoning` 支援：`none`、`low`、`medium`、`high`、`xhigh`。各 provider 是否支援與如何對應由 `core` 統一 clamp。

### `router.New`

```go
func New(config Config) (provider.Agent, error)
```

依 `config.Name` 的 provider 前綴建立對應 Agent。未知 provider 回傳 error。

### `router.Config`

| 欄位 | 用途 |
|------|------|
| `Name` | `<provider>@<model>` |
| `APIKey` | API Key 供應商使用 |
| `Token` | OAuth token（`*CopilotToken` / `*CodexToken` / `*GrokToken`） |
| `BaseURL` | 僅 `compat` |
| `AccountID` / `GatewayID` | 僅 `cloudflare` |

### `core.Config`

| 欄位 | 用途 |
|------|------|
| `Model` | 模型名稱（不含 provider 前綴） |
| `APIKey` | API Key |
| `Token` | OAuth token（any） |
| `BaseURL` | 自訂 endpoint |
| `AccountID` / `GatewayID` | Cloudflare |

### `core.Message` / `core.Tool` / `core.Output`

| 型別 | 重點欄位 |
|------|----------|
| `Message` | `Role`、`Content`、`ReasoningContent`、`ToolCalls`、`ToolCallID` |
| `ContentPart` | 多模態：`text` / `image_url` |
| `Tool` | OpenAI function-calling 形狀 |
| `Output` | `Choices`、`Usage`、可選 `Error` |
| `Usage` | `Input`、`Output`、`CacheCreate`、`CacheRead` |

### Reasoning 輔助函式

| 函式 | 說明 |
|------|------|
| `ClampReasoningLevel` | 將 level 上限限制在模型支援範圍 |
| `FloorReasoningLevel` | 將 level 下限抬到模型最低要求 |
| `MaxReasoningLevel` / `MinReasoningLevel` | 查詢模型上下限 |
| `SupportsReasoningSwitch` | 是否支援 reasoning 切換 |
| `SupportTemperature` | 是否支援 temperature |
| `ResponsesAPI` | 是否走 OpenAI Responses API |
| `GetThinkingType` / `GetThinkingConfig` / `ThinkingBudget` | Claude / Gemini thinking 參數 |

### OAuth 套件

| 路徑 | 用途 |
|------|------|
| `core/oauth/copilot` | GitHub device flow、keychain load/clear |
| `core/oauth/openaiCodex` | Codex OAuth |
| `core/oauth/grok` | Grok OAuth |

Token 型別內建 `Expired()`（Codex / Grok）供 refresh 判斷。

### Codex 圖片生成

`core/openaiCodex` 的 Agent 額外提供：

```go
func (a *Agent) GenerateImage(ctx context.Context, prompt string, opts ImageOptions) (string, string, error)
```

`ImageOptions` 支援 `Size`、`Quality`、參考圖 base64（`RefImageB64` / `RefMime`）。

***

©️ 2026 [邱敬幃 Pardn Chiu](https://pardn.io)
