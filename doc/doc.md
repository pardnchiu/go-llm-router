# go-llm-router - Documentation

> Back to [README](../README.md)

## Prerequisites

- Go 1.25 or later
- Credentials for the provider you will use: an API key, or an OAuth token for Copilot, Codex, or Grok OAuth
- Access to the operating-system keychain when using packages under `core/oauth`

## Installation

### Add the module

```bash
go get github.com/pardnchiu/go-llm-router
```

### Build from source

```bash
git clone https://github.com/pardnchiu/go-llm-router.git
cd go-llm-router
go build ./...
```

### Run the local OpenAI-compatible test server

`make test` runs `go run ./cmd/test`, which starts an HTTP server on port `8787` by default.

```bash
export OPENAI_API_KEY="your-api-key"
make test
```

Set `PORT` to use a different listener port:

```bash
PORT=8080 make test
```

## Configuration

### `router.Config`

Pass `router.Config` to `router.New` to select and configure a provider.

| Field | Required | Description |
|---|---:|---|
| `Name` | Yes | Provider and model in `provider@model` form |
| `APIKey` | Conditional | API credential for key-based providers |
| `Token` | Conditional | OAuth token object for `copilot`, `codex`, or `grok-oauth` |
| `BaseURL` | `compat` only | Base URL of an OpenAI-compatible endpoint |
| `AccountID` | `cloudflare` only | Cloudflare account identifier |
| `GatewayID` | `cloudflare` only | Cloudflare AI Gateway identifier |

Provider prefixes are `claude`, `openai`, `gemini`, `grok`, `grok-oauth`, `deepseek`, `nvidia`, `openrouter`, `cloudflare`, `compat`, `copilot`, and `codex`.

### Model-name formats

| Format | Meaning | Example |
|---|---|---|
| `<provider>@<model>` | Standard router key | `openai@gpt-5.4` |
| `<provider>[<tag>]@<model>` | The optional bracket tag is ignored while selecting the provider | `claude[eu]@claude-opus-4-8` |
| `compat@<model>` | Custom OpenAI-compatible endpoint; also requires `BaseURL` | `compat@my-local-model` |

### Credential requirements for the test server

The test server obtains credentials from environment variables for each request.

| Provider prefix | Required environment variables |
|---|---|
| `claude` | `ANTHROPIC_API_KEY` |
| `openai` | `OPENAI_API_KEY` |
| `gemini` | `GEMINI_API_KEY` |
| `grok` | `XAI_API_KEY` |
| `deepseek` | `DEEPSEEK_API_KEY` |
| `nvidia` | `NVIDIA_API_KEY` |
| `openrouter` | `OPENROUTER_API_KEY` |
| `cloudflare` | `CLOUDFLARE_API_KEY`, `CLOUDFLARE_ACCOUNT_ID`, `CLOUDFLARE_GATEWAY_ID` |
| `compat` | `COMPAT_API_KEY`, `COMPAT_BASE_URL` |
| `copilot` | `COPILOT_TOKEN` |

`COPILOT_TOKEN` accepts either a raw access token or the JSON token produced by the Copilot OAuth flow. The test server does not expose Codex or Grok OAuth model routing.

## Usage

### Basic request through the router

Create an agent, build messages, then choose both a reasoning level and an execution mode. `ModeDefault` is the portable choice; supported provider/model pairs can use `ModeFast`.

```go
package main

import (
	"context"
	"fmt"
	"log"

	core "github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/router"
)

func main() {
	agent, err := router.New(router.Config{
		Name:   "openai@gpt-5.4",
		APIKey: "your-api-key",
	})
	if err != nil {
		log.Fatal(err)
	}

	out, status, err := agent.Send(
		context.Background(),
		[]core.Message{{Role: "user", Content: "Explain Go interfaces in one sentence."}},
		nil,
		core.ReasoningMedium,
		core.ModeDefault,
	)
	if err != nil {
		log.Fatalf("request failed (HTTP %d): %v", status, err)
	}
	if len(out.Choices) == 0 {
		log.Fatal("provider returned no choices")
	}

	fmt.Println(out.Choices[0].Message.Content)
}
```

### Request fast service when supported

Fast mode is model-specific. Check it before selecting `ModeFast`; an unsupported model simply receives no provider-specific fast-tier request.

```go
mode := core.ModeDefault
if core.SupportFast("openai", "gpt-5.4") {
	mode = core.ModeFast
}

out, status, err := agent.Send(ctx, messages, nil, core.ReasoningHigh, mode)
if err != nil {
	return fmt.Errorf("HTTP %d: %w", status, err)
}
_ = out
```

Fast-capable routes use provider-native controls: Claude sets `speed: fast`; OpenAI sets `service_tier: fast`; Grok and supported OpenRouter routes set a priority tier. When a provider reports a downgrade, the library records a warning through `slog`.

### Stream output

Providers that implement `core.StreamAgent` can expose text, reasoning, tool-call, usage, completion, and error events through a channel.

```go
streamer, ok := agent.(core.StreamAgent)
if !ok {
	return fmt.Errorf("%s does not support streaming", agent.Name())
}

events, err := streamer.SendStream(ctx, messages, tools, core.ReasoningMedium, core.ModeDefault)
if err != nil {
	return err
}

for event := range events {
	switch event.Type {
	case core.StreamEventText:
		fmt.Print(event.TextDelta)
	case core.StreamEventReasoning:
		fmt.Print(event.ReasoningDelta)
	case core.StreamEventToolCall:
		fmt.Printf("tool delta: %+v\n", event.ToolCall)
	case core.StreamEventUsage:
		fmt.Printf("usage: %+v\n", event.Usage)
	case core.StreamEventError:
		return event.Err
	}
}
```

### Use function tools

Tools use the OpenAI function-calling shape. Provider adapters translate it into each upstream API's native schema.

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	core "github.com/pardnchiu/go-llm-router/core"
)

func request(ctx context.Context, agent core.Agent) error {
	tools := []core.Tool{{
		Type: "function",
		Function: core.ToolFunction{
			Name:        "get_weather",
			Description: "Return the weather for a city",
			Parameters: json.RawMessage(`{
				"type":"object",
				"properties":{"city":{"type":"string"}},
				"required":["city"]
			}`),
		},
	}}

	out, status, err := agent.Send(
		ctx,
		[]core.Message{{Role: "user", Content: "What is the weather in Taipei?"}},
		tools,
		core.ReasoningMedium,
		core.ModeDefault,
	)
	if err != nil {
		return fmt.Errorf("HTTP %d: %w", status, err)
	}
	if len(out.Choices) == 0 {
		return fmt.Errorf("provider returned no choices")
	}

	for _, call := range out.Choices[0].Message.ToolCalls {
		log.Printf("call %s(%s)", call.Function.Name, call.Function.Arguments)
	}
	return nil
}
```

After executing a tool, append the assistant tool-call message and a `tool` message that sets `ToolCallID` to the call ID before sending the next turn.

### Send multimodal content

`Message.Content` may be a string or a slice of `core.ContentPart`. Providers that accept data URLs can map an `image_url` part to their native image representation.

```go
messages := []core.Message{{
	Role: "user",
	Content: []core.ContentPart{
		{Type: "text", Text: "Describe this image."},
		{Type: "image_url", ImageURL: &core.ImageURL{
			URL: "data:image/png;base64,iVBORw0KGgoAAAANSUhEUg...",
		}},
	},
}}
```

### Load an OAuth token

OAuth packages persist tokens in the operating-system keychain, return `nil, nil` when no stored token exists, and refresh expired Codex or Grok tokens through `EnsureFresh` during requests.

```go
package main

import (
	"context"
	"fmt"
	"log"

	oauthCopilot "github.com/pardnchiu/go-llm-router/core/oauth/copilot"
	"github.com/pardnchiu/go-llm-router/core/router"
)

func main() {
	ctx := context.Background()
	token, err := oauthCopilot.Load()
	if err != nil {
		log.Fatal(err)
	}
	if token == nil {
		token, err = oauthCopilot.LoginWithCallback(ctx, func(code *oauthCopilot.DeviceCode) {
			fmt.Println("Open", code.VerificationURI, "and enter", code.UserCode)
		})
		if err != nil {
			log.Fatal(err)
		}
	}

	agent, err := router.New(router.Config{
		Name:  "copilot@gpt-5",
		Token: token,
	})
	if err != nil {
		log.Fatal(err)
	}
	_ = agent
}
```

### Use the OpenAI-compatible test endpoint

The local server accepts `POST /v1/chat/completions`, validates `model`, `reasoning`, and `mode`, then selects a configured agent. Set `stream` to receive SSE chunks.

```bash
curl http://127.0.0.1:8787/v1/chat/completions \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "openai@gpt-5.4",
    "messages": [{"role": "user", "content": "Hello"}],
    "reasoning": "medium",
    "mode": "default"
  }'
```

For streaming, add `"stream": true`; the endpoint forwards internal stream events as OpenAI-style `chat.completion.chunk` SSE records and ends with `data: [DONE]`.

## API Reference

### `core.Agent`

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
```

`Send` returns a normalized `Output`, the upstream HTTP status code, and an error. Every provider adapter implements this contract.

### `core.StreamAgent`

```go
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

| Event type | Payload |
|---|---|
| `StreamEventText` | `TextDelta` |
| `StreamEventReasoning` | `ReasoningDelta` |
| `StreamEventToolCall` | `ToolCall` delta with index, ID, name, or argument fragment |
| `StreamEventUsage` | Normalized `Usage` |
| `StreamEventDone` | `FinishReason` |
| `StreamEventError` | `Err` |

### Reasoning and mode

| Type | Values | Purpose |
|---|---|---|
| `Reasoning` | `none`, `low`, `medium`, `high`, `xhigh`, `max` | Requested reasoning effort; adapters clamp values to each model's supported range |
| `Mode` | `default`, `fast` | Execution tier request; use `SupportFast` to check model support |

`ParseReasoning` also accepts `minimal`, `extra`, and `ultra` aliases. `ParseMode` parses `default` and `fast`. `WarnFastDowngrade` logs a warning when an upstream response reports a non-fast tier.

### Core payload types

| Type | Key fields | Purpose |
|---|---|---|
| `Message` | `Role`, `Content`, `ReasoningContent`, `ToolCalls`, `ToolCallID` | Conversation input and tool-result linkage |
| `ContentPart` / `ImageURL` | `text` and `image_url` data | Text and image input parts |
| `Tool` / `ToolFunction` | `Name`, `Description`, JSON `Parameters` | Function-tool definition |
| `ToolCall` | `ID`, function `Name`, `Arguments`, `ThoughtSignature` | Provider-emitted call payload |
| `Output` / `OutputChoices` | `Choices`, `Usage`, optional `ServiceTier` / `Error` | Normalized completion response |
| `Usage` | `Input`, `Output`, `CacheCreate`, `CacheRead` | Normalized token accounting |

`Usage.UnmarshalJSON` combines common OpenAI- and Anthropic-style token fields and separates cache reads from billable input.

### Provider and router packages

| Package | Responsibility |
|---|---|
| `core/router` | Parses the provider prefix and creates the matching `core.Agent` |
| `core/claude` | Anthropic Messages API, prompt caching, fast mode, and streaming |
| `core/openai` | OpenAI Chat Completions or Responses API |
| `core/gemini` | Google Gemini content conversion, schema sanitization, and cache support |
| `core/grok` / `core/grokOauth` | xAI API-key and OAuth Responses API routes |
| `core/copilot` | GitHub Copilot Chat or Responses API selection and streaming |
| `core/openaiCodex` | ChatGPT Codex OAuth Responses API and image generation |
| `core/deepseek`, `core/nvidia`, `core/openRouter`, `core/cloudflare`, `core/compat` | Additional key-based provider adapters |
| `core/oauth/*` | Login, load, refresh, and keychain storage for OAuth credentials |

### Codex image generation

`core/openaiCodex.Agent` also provides image generation.

```go
func (a *Agent) GenerateImage(
	ctx context.Context,
	prompt string,
	opts ImageOptions,
) (base64Image string, revisedPrompt string, err error)
```

`ImageOptions` supports `Size`, `Quality`, and optional base64 reference-image fields.

***

©️ 2026 [邱敬幃 Pardn Chiu](https://pardn.io)
