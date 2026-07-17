# go-llm-router - Documentation

> Back to [README](../README.md)

## Prerequisites

- Go 1.25 or higher
- At least one provider credential: an API key, or an OAuth token for Copilot / Codex / Grok
- System keychain write access when using `core/oauth` token storage

## Installation

### go get

```bash
go get github.com/pardnchiu/go-llm-router
```

### From Source

```bash
git clone https://github.com/pardnchiu/go-llm-router.git
cd go-llm-router
go build ./...
```

## Configuration

### router.Config

| Field | Required | Description |
|-------|----------|-------------|
| `Name` | Yes | `provider@model`, for example `claude@claude-sonnet-5` |
| `APIKey` | Conditional | API key for `claude` / `openai` / `gemini` / `grok` / `deepseek` / `nvidia` / `openrouter` / `cloudflare` / `compat` |
| `Token` | Conditional | OAuth token for `copilot` / `codex` / `grok-oauth` |
| `BaseURL` | Conditional | Custom OpenAI-compatible endpoint for `compat` only |
| `AccountID` | Conditional | Cloudflare Account ID |
| `GatewayID` | Conditional | Cloudflare AI Gateway ID |

### provider.Config

| Field | Description |
|-------|-------------|
| `Model` | Model ID without the provider prefix |
| `APIKey` | Provider API key |
| `Token` | Provider-specific OAuth token object |
| `BaseURL` | Custom endpoint base URL |
| `AccountID` / `GatewayID` | Cloudflare-only fields |

### Name Formats

| Format | Description | Example |
|--------|-------------|---------|
| `<provider>@<model>` | Standard routing key | `openai@gpt-5.4` |
| `<provider>[tag]@<model>` | Bracket tags are ignored for matching | `claude[eu]@claude-opus-4-8` |
| `compat@<model>` | Custom endpoint; requires `BaseURL` | `compat@my-model` |

Supported provider keys: `claude`, `openai`, `gemini`, `grok`, `grok-oauth`, `deepseek`, `nvidia`, `openrouter`, `cloudflare`, `compat`, `copilot`, `codex`.

## Usage

### Basic

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
		{Role: "user", Content: "Hello"},
	}

	output, statusCode, err := agent.Send(context.Background(), messages, nil, "medium")
	if err != nil {
		panic(err)
	}

	fmt.Println(statusCode, output.Choices[0].Message.Content)
}
```

### Direct Provider Construction

```go
import (
	provider "github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/claude"
)

agent, err := claude.New(provider.Config{
	Model:  "claude-sonnet-5",
	APIKey: "sk-ant-...",
})
if err != nil {
	panic(err)
}
```

### Tool Calls

```go
tools := []provider.Tool{{
	Type: "function",
	Function: provider.ToolFunction{
		Name:        "get_weather",
		Description: "Look up weather for a city",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}`),
	},
}}

output, _, err := agent.Send(ctx, messages, tools, "medium")
if err != nil {
	panic(err)
}

for _, call := range output.Choices[0].Message.ToolCalls {
	fmt.Println(call.Function.Name, call.Function.Arguments)
}
```

### Token Usage

```go
output, _, err := agent.Send(ctx, messages, nil, "medium")
if err != nil {
	panic(err)
}

usage := output.Usage
fmt.Printf("input=%d output=%d cache_create=%d cache_read=%d\n",
	usage.Input, usage.Output, usage.CacheCreate, usage.CacheRead)
```

### OAuth Providers

```go
import (
	oauthCopilot "github.com/pardnchiu/go-llm-router/core/oauth/copilot"
	"github.com/pardnchiu/go-llm-router/core/router"
)

token, err := oauthCopilot.Load()
if err != nil {
	panic(err)
}
if token == nil {
	token, err = oauthCopilot.LoginWithCallback(ctx, func(code *oauthCopilot.DeviceCode) {
		fmt.Println(code.VerificationURI, code.UserCode)
	})
	if err != nil {
		panic(err)
	}
}
	agent, err := router.New(router.Config{
		Name:  "copilot@gpt-5",
		Token: token,
	})
```

### Image Generation (Codex)

```go
import openaicodex "github.com/pardnchiu/go-llm-router/core/openaiCodex"

agent, err := openaicodex.New(provider.Config{
	Model: "gpt-5",
	Token: codexToken,
})
if err != nil {
	panic(err)
}

b64, revised, err := agent.GenerateImage(ctx, "a cat in space", openaicodex.ImageOptions{
	Size:    "1024x1024",
	Quality: "high",
})
```

## API Reference

### Agent Interface

```go
type Agent interface {
	Name() string
	Send(ctx context.Context, messages []Message, toolDefs []Tool, reasoning string) (*Output, int, error)
}
```

`Send` performs one provider request. `reasoning` accepts `none` / `low` / `medium` / `high` / `xhigh` and is clamped per provider and model.

### router.New

```go
func New(config Config) (provider.Agent, error)
```

Parses `config.Name`, selects the matching factory, and returns a unified `Agent`.

### Core Types

| Type | Description |
|------|-------------|
| `Message` | Chat message with `Role`, `Content`, optional `ToolCalls` / `ToolCallID` / `ReasoningContent` |
| `Tool` / `ToolFunction` | OpenAI-style tool definition |
| `ToolCall` | Model-emitted tool call payload |
| `Output` / `OutputChoices` | Normalized response container |
| `Usage` | Normalized token counters: `Input`, `Output`, `CacheCreate`, `CacheRead` |
| `ContentPart` / `ImageURL` | Multimodal content parts |
| `CopilotToken` / `CodexToken` / `GrokToken` | OAuth token objects with expiry helpers |

### Reasoning Helpers

| Function | Description |
|----------|-------------|
| `SupportsReasoningSwitch` | Whether the model can switch reasoning levels |
| `MaxReasoningLevel` / `MinReasoningLevel` | Allowed upper and lower bounds |
| `ClampReasoningLevel` / `FloorReasoningLevel` | Bound a requested level into range |
| `SupportReasoningEffort` | Whether OpenAI-style `reasoning_effort` is used |
| `GetThinkingType` / `GetThinkingConfig` / `ThinkingBudget` | Claude / Gemini thinking mapping |
| `SupportTemperature` | Whether temperature is accepted |
| `ResponsesAPI` | Whether OpenAI Responses API should be used |

### Package Layout

| Package | Description |
|---------|-------------|
| `core` | Shared types, Agent interface, reasoning helpers |
| `core/router` | String-based Agent factory |
| `core/claude` ... `core/compat` | Provider implementations |
| `core/oauth/*` | OAuth login, load, refresh, and clear flows |

***

©️ 2026 [邱敬幃 Pardn Chiu](https://pardn.io)
