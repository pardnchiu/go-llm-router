# go-llm-router - Documentation

> Back to [README](../README.md)

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Reference](#api-reference)

## Prerequisites

- Go 1.25 or higher
- `github.com/pardnchiu/go-pkg` v0.13.14 (the only direct dependency, pulled in by `go get`)
- At least one provider credential:
  - Key-based: Anthropic, OpenAI, Gemini, xAI, DeepSeek, Mistral, NVIDIA, Ollama Cloud, OpenRouter, Cloudflare Workers AI
  - OAuth-based: a GitHub Copilot subscription, a ChatGPT (Codex) account, a Grok account
  - Any self-hosted or third-party OpenAI-compatible endpoint (`compat`)
- macOS Keychain or Linux `secret-tool` (OAuth providers only; tokens are stored through `go-pkg/filesystem/keychain`)

## Installation

### As a module dependency

```bash
go get github.com/pardnchiu/go-llm-router
```

### From source

```bash
git clone https://github.com/pardnchiu/go-llm-router.git
cd go-llm-router
go build ./...
```

## Configuration

### The library itself

The library reads no environment variables; callers pass every credential through `router.Config`. OAuth provider tokens live in the system keychain:

| Provider | Keychain key | How to obtain |
|---|---|---|
| Copilot | `COPILOT_OAUTH_TOKEN` | `core/oauth/copilot.LoginWithCallback` (device flow) |
| Codex | `CODEX_OAUTH_TOKEN` | `core/oauth/codex.LoginWithCallback` (PKCE browser authorization) |
| Grok | `GROK_OAUTH_TOKEN` | `core/oauth/grok.LoginWithCallback` (PKCE browser authorization) |

All three fall back to the legacy key `agenvoy.<provider>.token` on read.

## Usage

### Basic: resolve an Agent and send a request

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/router"
)

func main() {
	agent, err := router.New(router.Config{
		Name:   "claude@claude-opus-5",
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	})
	if err != nil {
		log.Fatalf("router.New: %v", err)
	}

	out, code, err := agent.Send(context.Background(),
		[]llmrouter.Message{{Role: "user", Content: "Summarize Go generics in one sentence."}},
		nil, llmrouter.ReasoningDefault, llmrouter.ModeDefault)
	if err != nil {
		log.Fatalf("Send: http %d: %v", code, err)
	}
	if len(out.Choices) == 0 {
		log.Fatal("no choices returned")
	}

	fmt.Println(out.Choices[0].Message.Content)
	fmt.Printf("input=%d output=%d cache_read=%d\n",
		out.Usage.Input, out.Usage.Output, out.Usage.CacheRead)
}
```

### Streaming

`SendStream` is an optional interface obtained by type assertion; agents without streaming fail the assertion.

```go
streamAgent, ok := agent.(llmrouter.StreamAgent)
if !ok {
	return fmt.Errorf("%s does not support streaming", agent.Name())
}

events, err := streamAgent.SendStream(ctx, messages, nil, llmrouter.ReasoningHigh, llmrouter.ModeDefault)
if err != nil {
	var streamErr *llmrouter.StreamError
	if errors.As(err, &streamErr) {
		return fmt.Errorf("%s: http %d: %s", streamErr.Provider, streamErr.Code, streamErr.Body)
	}
	if errors.Is(err, llmrouter.ErrStreamUnsupported) {
		return fmt.Errorf("upstream returned a non-SSE response")
	}
	return err
}

for evt := range events {
	switch evt.Type {
	case llmrouter.StreamEventText:
		fmt.Print(evt.TextDelta)
	case llmrouter.StreamEventReasoning:
		fmt.Print(evt.ReasoningDelta)
	case llmrouter.StreamEventToolCall:
		fmt.Printf("\n[tool] %s %s\n", evt.ToolCall.Name, evt.ToolCall.Arguments)
	case llmrouter.StreamEventUsage:
		fmt.Printf("\n[usage] in=%d out=%d\n", evt.Usage.Input, evt.Usage.Output)
	case llmrouter.StreamEventError:
		return evt.Err
	case llmrouter.StreamEventDone:
		fmt.Printf("\n[done] %s\n", evt.FinishReason)
	}
}
```

### Reasoning levels and fast mode

```go
reasoning, ok := llmrouter.ParseReasoning("xhigh") // none / low / medium / high / xhigh / max
if !ok {
	reasoning = llmrouter.ReasoningDefault          // medium
}

// Ask the model for its supported range; out-of-range requests are clamped
if limited, ok := agent.(llmrouter.ReasoningAgent); ok {
	low, high := limited.ReasoningLimits()
	reasoning = llmrouter.ClampReasoning(reasoning, low, high, "claude", "claude-opus-5")
}

mode := llmrouter.ModeDefault
if llmrouter.SupportFast("claude", "claude-opus-5") {
	mode = llmrouter.ModeFast
}

out, _, err := agent.Send(ctx, messages, nil, reasoning, mode)
```

### Tool calls

```go
params := json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}`)

tools := []llmrouter.Tool{{
	Type: "function",
	Function: llmrouter.ToolFunction{
		Name:        "get_weather",
		Description: "Get the current weather for a city",
		Parameters:  params,
	},
}}

out, _, err := agent.Send(ctx, messages, tools, llmrouter.ReasoningDefault, llmrouter.ModeDefault)
if err != nil {
	return err
}
for _, call := range out.Choices[0].Message.ToolCalls {
	fmt.Println(call.Function.Name, call.Function.Arguments)
}
```

Return tool results with `Role: "tool"` plus `ToolCallID`. Gemini's `ThoughtSignature` must be echoed back verbatim or the next turn is rejected.

### Multimodal input

```go
messages := []llmrouter.Message{{
	Role: "user",
	Content: []llmrouter.ContentPart{
		{Type: "text", Text: "Describe this image."},
		{Type: "image_url", ImageURL: &llmrouter.ImageURL{
			URL: llmrouter.DataURI("image/png", base64.StdEncoding.EncodeToString(raw)),
		}},
	},
}}
```

### Listing models

```go
ids, err := gemini.Models(ctx, llmrouter.Config{APIKey: key}, llmrouter.ModelFilter{TextOnly: true})
if err != nil {
	return err
}

// Image models only; video models never slip through
images, err := gemini.Models(ctx, llmrouter.Config{APIKey: key}, llmrouter.ModelFilter{ImageOnly: true})

// gemini / mistral / copilot also expose ModelInfos with thinking, efforts, endpoints
infos, err := gemini.ModelInfos(ctx, llmrouter.Config{APIKey: key}, llmrouter.ModelFilter{TextOnly: true})
```

### Image generation

The image model is the agent model, picked the same way as chat and TTS models. `codex` is the exception: its backend generates from the chat model itself.

```go
agent, err := router.New(router.Config{Name: "openai@gpt-image-2", APIKey: key})
if err != nil {
	return err
}

imageAgent, ok := agent.(llmrouter.ImageAgent)
if !ok {
	return fmt.Errorf("%s cannot generate images", agent.Name())
}

result, err := imageAgent.GenerateImage(ctx, "A lighthouse at dawn, watercolor", llmrouter.ImageOptions{
	AspectRatio: "16:9",
	Size:        "1k",
	Quality:     "high",
})
if err != nil {
	return err
}
os.WriteFile("out.png", mustDecode(result.B64), 0644)
```

### Speech to text and text to speech

```go
sttAgent, ok := agent.(llmrouter.STTAgent)
if !ok {
	return fmt.Errorf("%s cannot transcribe", agent.Name())
}
text, err := sttAgent.Transcribe(ctx, raw, llmrouter.STTOptions{MimeType: "audio/mp3", Language: "zh"})
if err != nil {
	return err
}

ttsAgent, ok := agent.(llmrouter.TTSAgent)
if !ok {
	return fmt.Errorf("%s cannot speak", agent.Name())
}
speech, err := ttsAgent.Speak(ctx, text.Text, llmrouter.TTSOptions{Voice: "alloy", Format: "mp3"})
if err != nil {
	return err
}
os.WriteFile("out."+strings.TrimPrefix(speech.MimeType, "audio/"), speech.Audio, 0644)
```

### Querying balance and quota

```go
remaining, err := openrouter.Usage(ctx, llmrouter.Config{APIKey: key})
if err != nil {
	return err
}
fmt.Printf("credits left: %.2f\n", remaining)
```

### Self-hosted and third-party compatible endpoints

```go
// Explicit compat instance
agent, err := router.New(router.Config{
	Name:    "compat[lmstudio]@qwen3-30b",
	BaseURL: "http://127.0.0.1:1234/v1",
	APIKey:  "not-needed",
})

// An unknown prefix falls through to compat; Name() reports "groq@llama-3.3-70b"
agent, err = router.New(router.Config{
	Name:    "groq@llama-3.3-70b",
	BaseURL: "https://api.groq.com/openai/v1",
	APIKey:  os.Getenv("GROQ_API_KEY"),
})
```

An empty `BaseURL` defaults to `http://localhost:11434/v1` (Ollama).

## API Reference

### Interfaces

```go
type Agent interface {
	Name() string
	Send(ctx context.Context, messages []Message, toolDefs []Tool, reasoning Reasoning, mode Mode) (*Output, int, error)
}

type StreamAgent interface {
	SendStream(ctx context.Context, messages []Message, toolDefs []Tool, reasoning Reasoning, mode Mode) (<-chan StreamEvent, error)
}

type ReasoningAgent interface {
	ReasoningLimits() (min, max Reasoning)
}

type ImageAgent interface {
	GenerateImage(ctx context.Context, prompt string, opts ImageOptions) (*ImageResult, error)
}

type STTAgent interface {
	Transcribe(ctx context.Context, audio []byte, opts STTOptions) (*STTResult, error)
}

type TTSAgent interface {
	Speak(ctx context.Context, text string, opts TTSOptions) (*TTSResult, error)
}
```

`Agent` is mandatory; every other interface is reached by type assertion.

The import path is `github.com/pardnchiu/go-llm-router/core`; the package it declares is `llmrouter`, so symbols read as `llmrouter.Agent`.

### router

| Symbol | Signature | Description |
|---|---|---|
| `router.Config` | `struct{ Name, APIKey string; Token any; BaseURL, AccountID, GatewayID string }` | `Name` is `provider@model`; `Token` carries `*llmrouter.CopilotToken` / `*llmrouter.CodexToken` / `*llmrouter.GrokToken` for OAuth providers |
| `router.New` | `func(config Config) (llmrouter.Agent, error)` | Builds the Agent for the prefix; an unknown prefix containing `@` becomes `compat[<prefix>]@<model>` |

### Provider prefixes

| Prefix | Package | Credential | Streaming | Extras |
|---|---|---|---|---|
| `claude@` | `core/claude` | `APIKey` | ✓ | — |
| `openai@` | `core/openai` | `APIKey` | ✓ | Image, STT, TTS |
| `gemini@` | `core/gemini` | `APIKey` | ✓ | Image, STT, TTS, `cachedContents` prefix cache |
| `grok@` | `core/grok` | `APIKey` | ✓ | Image |
| `deepseek@` | `core/deepseek` | `APIKey` | ✓ | `Usage` |
| `mistral@` | `core/mistral` | `APIKey` | ✓ | `ModelInfos` |
| `nvidia@` | `core/nvidia` | `APIKey` | ✓ | — |
| `ollama-cloud@` | `core/ollamaCloud` | `APIKey` | ✓ | `Usage` |
| `openrouter@` | `core/openRouter` | `APIKey` | ✓ | Image, STT, TTS, `Usage` |
| `cloudflare@` | `core/cloudflare` | `APIKey` + `AccountID` + `GatewayID` | ✓ | — |
| `compat@` / `compat[name]@` | `core/compat` | `APIKey` + `BaseURL` | ✓ | — |
| `copilot@` | `core/copilot` | `*llmrouter.CopilotToken` | ✓ | `Usage`, `ModelInfos` |
| `codex@` | `core/openaiCodex` | `*llmrouter.CodexToken` | ✓ | Image, `Usage` |
| `grok-oauth@` | `core/grokOauth` | `*llmrouter.GrokToken` | ✓ | Image, `Usage` |

Every provider package exposes `New(llmrouter.Config) (*Agent, error)` and `Models(ctx, llmrouter.Config, llmrouter.ModelFilter) ([]string, error)`.

### Reasoning levels

| Constant | String | Alias |
|---|---|---|
| `ReasoningNone` | `none` | — |
| `ReasoningLow` | `low` | `minimal` |
| `ReasoningMedium` | `medium` (`ReasoningDefault`) | — |
| `ReasoningHigh` | `high` | — |
| `ReasoningXHigh` | `xhigh` | `extra` |
| `ReasoningMax` | `max` | `ultra` |

| Function | Signature | Description |
|---|---|---|
| `ParseReasoning` | `func(s string) (Reasoning, bool)` | Returns `ReasoningDefault, false` when unparseable |
| `ClampReasoning` | `func(r, lo, hi Reasoning, provider, model string) Reasoning` | Emits a debug log whenever it clamps |
| `OpenAIEffortRange` | `func(model string) (low, high Reasoning)` | Effort range per OpenAI model generation |

### Modes

| Symbol | Signature | Description |
|---|---|---|
| `ModeDefault` / `ModeFast` | `Mode` | `fast` maps onto each vendor's priority tier |
| `ParseMode` | `func(s string) (Mode, bool)` | Accepts `default` / `fast` |
| `SupportFast` | `func(providerName, model string) bool` | Decides whether a model generation offers a fast tier |
| `WarnFastDowngrade` | `func(providerName, model, tier string)` | Warns when the served tier is neither `fast` nor `priority` |

### Model filters

| Symbol | Signature | Description |
|---|---|---|
| `ModelFilter` | `struct{ TextOnly, STTOnly, TTSOnly, ImageOnly bool }` | Multiple flags intersect |
| `IsTextModel` / `IsSTTModel` / `IsTTSModel` / `IsImageModel` | `func(id string) bool` | Marker-based classification; `IsImageModel` excludes video models |
| `MatchModelFilter` | `func(id string, filter ModelFilter) bool` | Used by the `openai`, `gemini`, `grok`, and `grok-oauth` listings |
| `ModelInfo` | `struct{ ID string; Thinking bool; Efforts, Endpoints []string }` | Element type returned by `ModelInfos` |

Flag support differs per provider; an unsupported flag is ignored rather than rejected:

| Provider | `TextOnly` | `STTOnly` | `TTSOnly` | `ImageOnly` | Source |
|---|---|---|---|---|---|
| `openai`, `gemini`, `grok`, `grok-oauth` | ✓ | ✓ | ✓ | ✓ | model-ID markers |
| `openrouter` | ✓ | ✓ | ✓ | ✓ | `/api/v1/images/models`, `/api/v1/models?output_modalities=speech` / `transcription` |
| `cloudflare` | ✓ | ✓ | ✓ | — | Workers AI task name |
| `claude`, `copilot`, `deepseek`, `mistral`, `nvidia`, `ollama-cloud`, `codex` | ✓ | — | — | — | model-ID markers |

### Messages and output

| Type | Description |
|---|---|
| `Message` | `Role` / `Content` (`string` or `[]ContentPart`) / `ReasoningContent` / `ToolCalls` / `ToolCallID` |
| `ContentPart`, `ImageURL` | `text` and `image_url` parts; `image_url` accepts a data URI |
| `Tool`, `ToolFunction` | OpenAI-shaped tool definitions with `Parameters` as `json.RawMessage` |
| `ToolCall` | Carries Gemini's `ThoughtSignature` |
| `Output`, `OutputChoices` | `Choices` / `Usage` / `ServiceTier` / `Error` |
| `Usage` | `Input` / `Output` / `CacheCreate` / `CacheRead`, with a custom `UnmarshalJSON` that absorbs each vendor's field names |

`Usage.UnmarshalJSON` adds `prompt_tokens` to `input_tokens` and `completion_tokens` to `output_tokens`, then subtracts `prompt_tokens_details.cached_tokens` from `Input` and folds it into `CacheRead` so cache hits are not billed twice.

### Streaming

| Symbol | Description |
|---|---|
| `StreamEvent` | `Type` is `text` / `reasoning` / `tool_call` / `usage` / `done` / `error` |
| `ToolCallDelta` | Streamed tool-call increment with `Index` and `ThoughtSignature` |
| `StreamError` | Carries `Provider` / `Code` / `Body`; `errors.As` recovers the upstream status |
| `ErrStreamUnsupported` | Wrapped in `StreamError.Err` when the upstream answers with non-SSE content |
| `OpenStream` | `func(ctx, client, url, headers, body, label) (*http.Response, error)` |
| `StreamChat` / `StreamResponses` | Parse Chat Completions and Responses API SSE into an event channel |
| `ScanSSE` | `func(reader *bufio.Reader, handle func(event, data string) bool) error` |
| `StreamBodyLimit` / `ErrorBodyLimit` / `ErrorFrameLimit` / `JSONBodyLimit` | Read caps of 64 MiB / 8 KiB / 512 B / 64 KiB |

### Images

```go
type ImageOptions struct {
	AspectRatio string // "1:1" "16:9" "4:3"
	Size        string // "1k" "2k" "4k"
	Quality     string // "low" "medium" "high"
	RefImageB64 string
	RefMime     string
}

type ImageResult struct {
	B64      string
	MimeType string
	Revised  string
}
```

| Provider | Image model | Endpoint | Result |
|---|---|---|---|
| `openai` | agent model, e.g. `openai@gpt-image-2` | `/v1/images/generations`, `/v1/images/edits` with a reference | PNG (`output_format`), `Revised` set |
| `codex` | backend default | ChatGPT Codex Responses, `image_generation` tool | PNG, `Revised` set |
| `gemini` | agent model, e.g. `gemini@gemini-3.1-flash-image` | `:generateContent` | JPEG |
| `grok` / `grok-oauth` | agent model, e.g. `grok@grok-imagine-image-2.0` | `/v1/images/generations` and `/v1/images/edits` | JPEG |
| `openrouter` | agent model, e.g. `openrouter@google/gemini-2.5-flash-image` | `/api/v1/images` | `media_type` from the response |

| Option | `openai` | `codex` | `gemini` | `grok` / `grok-oauth` | `openrouter` |
|---|---|---|---|---|---|
| `AspectRatio` + `Size` | `size` as `WIDTHxHEIGHT` | ignored by the endpoint | `imageConfig.aspectRatio` / `.imageSize` | `aspect_ratio` / `resolution`, `4k` clamped to `2k` | `aspect_ratio` / `resolution` |
| `Quality` | `quality` | ignored by the endpoint | no counterpart | `quality` | `quality` |
| `RefImageB64` | `image` file field on `/v1/images/edits` | `input_image` part | `inline_data` part | `image.url` on `/v1/images/edits` | `input_references` data URL |

`ImagePixelSize` applies `Size` to the **short** edge: `16:9` + `1k` becomes `1824x1024`. Scaling the long edge instead falls below the endpoint's minimum pixel budget and is rejected. The ChatGPT Codex backend always answers `1254x1254` at `quality: low`, so `codex` forwards neither option.

### Audio

```go
type STTOptions struct{ Prompt, Language, MimeType string }
type TTSOptions struct{ Voice, Format string } // wav (default) / mp3 / opus / aac / flac / pcm
```

| Provider | STT | TTS | Notes |
|---|---|---|---|
| `openai` | `/v1/audio/transcriptions` | `/v1/audio/speech` | Model is the agent model; default voice `alloy` |
| `openrouter` | multipart `/api/v1/audio/transcriptions` | `/api/v1/audio/speech` | An empty `Voice` resolves to the first `supported_voices` entry of the model in the OpenRouter catalog; a voice outside that list fails locally with the valid list; only `mp3` passes through, every other `Format` requests `pcm` and returns WAV via `WrapPCM16` |
| `gemini` | verbatim transcript via `:generateContent` | `AUDIO` modality of `:generateContent` | Default voice `Kore`; PCM is wrapped into WAV by `WrapPCM16` |

| Function | Signature | Description |
|---|---|---|
| `PCMRate` | `func(mime string) int` | Reads the rate from `audio/L16;rate=24000`, falling back to `24000` |
| `WrapPCM16` | `func(pcm []byte, rate int) []byte` | Prepends a 44-byte WAV header |
| `DataURI` | `func(mime, b64 string) string` | Defaults to `image/png` when mime is empty |

### OAuth

`core/oauth/copilot`, `core/oauth/codex`, and `core/oauth/grok` expose the same shape:

| Function | Signature | Description |
|---|---|---|
| `Load` | `func() (*llmrouter.XxxToken, error)` | Reads and decodes the token from the keychain |
| `HasToken` | `func() bool` | Reports whether a login exists |
| `ClearToken` | `func() error` | Removes the token, legacy key included |
| `LoginWithCallback` | `func(ctx, onCode/onURL) (*llmrouter.XxxToken, error)` | Device flow for Copilot (yields `*DeviceCode`), PKCE authorization URL for Codex and Grok |
| `EnsureFresh` / `EnsureFreshSession` | `func(ctx, token[, refresh]) (...)` | Refreshes 60 seconds before expiry; Copilot additionally exchanges a short-lived session token |

### Usage queries

| Package | Signature | Return semantics |
|---|---|---|
| `core/openRouter` | `Usage(ctx, llmrouter.Config) (float64, error)` | Remaining credits (`total_credits - total_usage`) |
| `core/deepseek` | same | Account balance |
| `core/ollamaCloud` | same | Remaining monthly quota as a percentage |
| `core/copilot` | same (needs `Token`) | Remaining percentage of chat or premium interactions |
| `core/openaiCodex` | same (needs `APIKey`, `AccountID`) | Remaining percentage of the tighter rate-limit window |
| `core/grokOauth` | same | Remaining credit percentage |

### Miscellaneous

| Symbol | Signature | Description |
|---|---|---|
| `NewHTTPClient` | `func() *http.Client` | 10-minute timeout; `codex` and `grok-oauth` build their own client with a 15-second response-header timeout |
| `SupportTemperature` | `func(providerName, model string) bool` | Claude, the `gpt-5` family, `deepseek-reasoner`, and Gemini previews reject `temperature` |
| `ResponsesAPI` | `func(providerName, model string) bool` | Chooses Responses over Chat Completions for OpenAI and Copilot |
| `TruncateFrame` | `func(data string) string` | Caps an SSE frame in error messages at 512 bytes |
| `summary.VoiceReply` | `func(ctx, text string) (string, error)` | Compresses text over 320 characters into a speakable reply via Gemini, using `GEMINI_API_KEY` from the keychain |

***

©️ 2026 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
