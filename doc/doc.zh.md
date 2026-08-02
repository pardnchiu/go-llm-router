# go-llm-router - 技術文件

> 返回 [README](./README.zh.md)

## 前置需求

- Go 1.25.0 或更高版本
- 至少一組可用的供應商憑證：API Key，或 Copilot、Codex、Grok OAuth token
- 使用 OAuth token 持久化時，系統 keychain 必須可用

## 安裝

### 作為 Go 依賴使用

```bash
go get github.com/pardnchiu/go-llm-router
```

### 從原始碼驗證建置

```bash
git clone https://github.com/pardnchiu/go-llm-router.git
cd go-llm-router
go build ./...
```

## 設定

### `router.Config`

以 `router.New` 建立 Agent 時，使用 `Name` 選擇供應商與模型。

| 欄位 | 必填 | 說明 |
|---|---:|---|
| `Name` | 是 | `<provider>@<model>`，例如 `openai@gpt-5.4` |
| `APIKey` | 視供應商而定 | API Key 供應商的憑證 |
| `Token` | 視供應商而定 | Copilot、Codex 或 Grok OAuth 的 token 物件 |
| `BaseURL` | 僅 `compat` | OpenAI-compatible 服務的基底 URL |
| `AccountID`、`GatewayID` | 僅 `cloudflare` | Cloudflare Workers AI 與 AI Gateway 設定 |

| Name 前綴 | 驗證方式 |
|---|---|
| `claude`、`openai`、`gemini`、`grok`、`deepseek`、`nvidia`、`openrouter` | `APIKey` |
| `cloudflare` | `APIKey`、`AccountID`、`GatewayID` |
| `compat` | `BaseURL`；`APIKey` 可選 |
| `copilot`、`codex`、`grok-oauth` | `Token` |

`router.New` 會忽略 provider 名稱中 `[...]` 的附加標籤，因此 `claude[eu]@claude-opus-4-8` 仍會路由到 Claude 實作。

### 推理層級與執行模式

`Agent.Send` 與 `StreamAgent.SendStream` 都接收 `core.Reasoning` 和 `core.Mode`，請使用列舉值而非字串。

| `core.Reasoning` | 用途 |
|---|---|
| `ReasoningNone`、`ReasoningLow`、`ReasoningMedium` | 低到中等推理強度 |
| `ReasoningHigh`、`ReasoningXHigh`、`ReasoningMax` | 高強度推理；實際可用範圍由 provider 與模型限制 |

| `core.Mode` | 用途 |
|---|---|
| `ModeDefault` | 一般服務層級 |
| `ModeFast` | 請求支援快速層級的模型；不支援時會維持一般模式 |

可用 `core.ParseReasoning`、`core.ParseMode` 解析 API 輸入；`core.SupportFast(provider, model)` 可在送出前檢查快速模式資格。各 provider 會自行限制不支援的推理等級，並在快速服務層級被降級時記錄警告。

## 使用方式

### 基本：透過 router 建立 Agent

```go
package main

import (
	"context"
	"fmt"

	core "github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/router"
)

func main() {
	agent, err := router.New(router.Config{
		Name:   "openai@gpt-5.4",
		APIKey: "your-api-key",
	})
	if err != nil {
		panic(err)
	}

	output, statusCode, err := agent.Send(
		context.Background(),
		[]core.Message{{Role: "user", Content: "請用一句話說明 Go"}},
		nil,
		core.ReasoningMedium,
		core.ModeDefault,
	)
	if err != nil {
		panic(err)
	}
	if len(output.Choices) == 0 {
		panic("provider returned no choices")
	}

	fmt.Println(statusCode, output.Choices[0].Message.Content)
}
```

### 工具呼叫

工具以 OpenAI-style JSON Schema 表示；回應中的 `ToolCalls` 包含函式名稱與 JSON arguments。

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"

	core "github.com/pardnchiu/go-llm-router/core"
)

func run(ctx context.Context, agent core.Agent) error {
	tools := []core.Tool{{
		Type: "function",
		Function: core.ToolFunction{
			Name:        "get_weather",
			Description: "取得指定城市的天氣",
			Parameters: json.RawMessage(`{
				"type":"object",
				"properties":{"city":{"type":"string"}},
				"required":["city"]
			}`),
		},
	}}

	output, _, err := agent.Send(
		ctx,
		[]core.Message{{Role: "user", Content: "台北天氣如何？"}},
		tools,
		core.ReasoningMedium,
		core.ModeDefault,
	)
	if err != nil {
		return err
	}
	for _, call := range output.Choices[0].Message.ToolCalls {
		fmt.Println(call.Function.Name, call.Function.Arguments)
	}
	return nil
}
```

把工具執行結果以 `Role: "tool"` 與相對應的 `ToolCallID` 放回後續訊息，即可完成 tool-calling 迴圈。

### 串流回應

支援串流的 Agent 會實作 `core.StreamAgent`。先做型別斷言，再逐一處理文字、推理、工具呼叫、usage 與完成事件。

```go
package main

import (
	"context"
	"fmt"

	core "github.com/pardnchiu/go-llm-router/core"
)

func stream(ctx context.Context, agent core.Agent) error {
	streamAgent, ok := agent.(core.StreamAgent)
	if !ok {
		return fmt.Errorf("provider does not support streaming")
	}

	events, err := streamAgent.SendStream(
		ctx,
		[]core.Message{{Role: "user", Content: "寫一首短詩"}},
		nil,
		core.ReasoningMedium,
		core.ModeDefault,
	)
	if err != nil {
		return err
	}
	for event := range events {
		switch event.Type {
		case core.StreamEventText:
			fmt.Print(event.TextDelta)
		case core.StreamEventError:
			return event.Err
		}
	}
	return nil
}
```

### OAuth provider

OAuth 流程位於 `core/oauth/codex`、`core/oauth/copilot` 與 `core/oauth/grok`。各套件提供 token 載入、登入、更新與清除功能，並將 token 儲存在系統 keychain。取得 token 後，將它交給 `router.Config.Token`。

```go
package main

import (
	"context"
	"fmt"

	oauthCopilot "github.com/pardnchiu/go-llm-router/core/oauth/copilot"
	"github.com/pardnchiu/go-llm-router/core/router"
)

func copilotAgent(ctx context.Context) error {
	token, err := oauthCopilot.Load()
	if err != nil {
		return err
	}
	if token == nil {
		token, err = oauthCopilot.LoginWithCallback(ctx, func(code *oauthCopilot.DeviceCode) {
			fmt.Println("請開啟", code.VerificationURI, "並輸入", code.UserCode)
		})
		if err != nil {
			return err
		}
	}

	_, err = router.New(router.Config{Name: "copilot@gpt-5", Token: token})
	return err
}
```

### 回應與 token 使用量

`Send` 的第二個回傳值是 HTTP status code。標準化 usage 位於 `output.Usage`：

```go
fmt.Printf(
	"input=%d output=%d cache_create=%d cache_read=%d\n",
	output.Usage.Input,
	output.Usage.Output,
	output.Usage.CacheCreate,
	output.Usage.CacheRead,
)
```

### 本機相容測試服務

專案提供 OpenAI-compatible 的測試服務：

```bash
make test
```

它預設監聽 `:8787`（可用 `PORT` 覆寫），並提供 `POST /v1/chat/completions`。request 的 `model` 為 `<provider>@<model>`，可選 `reasoning`、`mode`、`tools` 與 `stream`。服務依 provider 讀取下列環境變數：

| Provider | 環境變數 |
|---|---|
| `claude` | `ANTHROPIC_API_KEY` |
| `openai` | `OPENAI_API_KEY` |
| `gemini` | `GEMINI_API_KEY` |
| `grok` | `XAI_API_KEY` |
| `deepseek` | `DEEPSEEK_API_KEY` |
| `nvidia` | `NVIDIA_API_KEY` |
| `openrouter` | `OPENROUTER_API_KEY` |
| `cloudflare` | `CLOUDFLARE_API_KEY`、`CLOUDFLARE_ACCOUNT_ID`、`CLOUDFLARE_GATEWAY_ID` |
| `compat` | `COMPAT_API_KEY`、`COMPAT_BASE_URL` |
| `copilot` | `COPILOT_TOKEN`（raw token 或 OAuth token JSON） |

啟用 `stream: true` 時，服務會將內部 `StreamEvent` 轉成 SSE `chat.completion.chunk`，最後送出 `data: [DONE]`。

## API 參考

### `core.Agent` 與 `core.StreamAgent`

```go
type Agent interface {
	Name() string
	Send(
		ctx context.Context,
		messages []Message,
		toolDefs []Tool,
		reasoning Reasoning,
		mode Mode,
	) (*Output, int, error)
}

type StreamAgent interface {
	SendStream(
		ctx context.Context,
		messages []Message,
		toolDefs []Tool,
		reasoning Reasoning,
		mode Mode,
	) (<-chan StreamEvent, error)
}
```

| 型別 | 重點 |
|---|---|
| `Message` | `Role`、文字或多模態 `Content`、`ReasoningContent`、`ToolCalls`、`ToolCallID` |
| `ContentPart` / `ImageURL` | `text` 與 `image_url` 內容部分 |
| `Tool` / `ToolFunction` | 函式名稱、說明與 JSON Schema parameters |
| `ToolCall` | 模型輸出的 function 名稱、arguments 與識別碼 |
| `Output` / `OutputChoices` | 標準化回應、finish reason 與可選 provider error |
| `Usage` | `Input`、`Output`、`CacheCreate`、`CacheRead` |
| `StreamEvent` | `text`、`reasoning`、`tool_call`、`usage`、`done`、`error` 事件 |

### `router.New`

```go
func New(config Config) (core.Agent, error)
```

解析 `Config.Name`，選擇對應供應商建構子，並回傳統一的 `core.Agent`。未知 provider 會回傳 error。

### Provider 差異

| Provider 群組 | 實作重點 |
|---|---|
| Claude | Anthropic Messages API、prompt caching、多模態與串流；符合資格時以 `speed: fast` 請求快速模式 |
| OpenAI | 依模型選擇 Chat Completions 或 Responses API；快速模式使用 `service_tier: fast` |
| Copilot | 查詢並快取模型 endpoint 能力，再選擇 Chat 或 Responses API；支援串流 |
| Gemini | Google GenerateContent、system instruction、function declaration、cache 與 thinking 設定 |
| Grok / Grok OAuth | API Key 與 OAuth Responses 路徑；快速模式請求 `service_tier: priority` |
| Codex | ChatGPT Codex Responses API、OAuth、SSE 組裝與 prompt cache key；另提供圖片生成方法 |
| DeepSeek、NVIDIA、OpenRouter、Cloudflare、compat | 各自的 Chat/Workers AI 或 OpenAI-compatible request 轉換與標準化回應 |

***

©️ 2026 [邱敬幃 Pardn Chiu](https://pardn.io)
