# Interfaces and Router

The Agent interfaces, the `router` factory, and what each provider package exports. Types live in `github.com/pardnchiu/go-llm-router/core`, which declares package `llmrouter`; the factory lives in `core/router`.

## Interfaces

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

## router

| Symbol | Signature | Description |
|---|---|---|
| `router.Config` | `struct{ Name, APIKey string; Token any; BaseURL, AccountID, GatewayID string; EnableClaude bool; StateDir string }` | `Name` is `provider@model`; `EnableClaude` / `StateDir` apply to `claude-code@` only |
| `router.New` | `func(config Config) (llmrouter.Agent, error)` | Builds the agent for the prefix; an unknown prefix containing `@` becomes `compat[<prefix>]@<model>` |

## Provider names

`core` declares one constant per provider prefix, without the trailing `@`. Build a name as `llmrouter.PROVIDER_GEMINI + "@" + model`; `Providers()` returns every prefix except `compat`.

| Constant | Value |
|---|---|
| `PROVIDER_CLAUDE` | `claude` |
| `PROVIDER_CLAUDE_CODE` | `claude-code` |
| `PROVIDER_OPENAI` | `openai` |
| `PROVIDER_GEMINI` | `gemini` |
| `PROVIDER_GROK` | `grok` |
| `PROVIDER_GROK_OAUTH` | `grok-oauth` |
| `PROVIDER_DEEPSEEK` | `deepseek` |
| `PROVIDER_MISTRAL` | `mistral` |
| `PROVIDER_NVIDIA` | `nvidia` |
| `PROVIDER_OLLAMA_CLOUD` | `ollama-cloud` |
| `PROVIDER_OPENROUTER` | `openrouter` |
| `PROVIDER_CLOUDFLARE` | `cloudflare` |
| `PROVIDER_COMPAT` | `compat` |
| `PROVIDER_COPILOT` | `copilot` |
| `PROVIDER_CODEX` | `codex` |

## Provider packages

Every provider package exposes `New(llmrouter.Config) (*Agent, error)`, `(*Agent).Name()`, and `Models(ctx, llmrouter.Config, llmrouter.ModelFilter) ([]string, error)`. `core/compat` has no `Models`, and `core/claudeCode.Models(ctx, llmrouter.Config)` takes no filter.

| Package | Prefix | Additional exports |
|---|---|---|
| `core/claude` | `claude@` | Messages API wire types `Output`, `Content` |
| `core/claudeCode` | `claude-code@` | `Usage`, `CheckBinary`, `Is` |
| `core/openai` | `openai@` | `ImageInput`, `ReadImageStream`, `ImageInstructions` |
| `core/gemini` | `gemini@` | `ModelInfos`; `:generateContent` wire types `Content`, `Part`, `InlineData`, `FunctionCall`, `FunctionResponse`, `Output` |
| `core/grok` | `grok@` | `RequestImage` |
| `core/grokOauth` | `grok-oauth@` | `Usage` |
| `core/deepseek` | `deepseek@` | `Usage` |
| `core/mistral` | `mistral@` | `ModelInfos` |
| `core/nvidia` | `nvidia@` | — |
| `core/ollamaCloud` | `ollama-cloud@` | — |
| `core/openRouter` | `openrouter@` | `Usage` |
| `core/cloudflare` | `cloudflare@` | — |
| `core/compat` | `compat@` / `compat[name]@` | — |
| `core/copilot` | `copilot@` | `Usage`, `ModelInfos` |
| `core/openaiCodex` | `codex@` | `Usage` |

## Shared wire helpers

Packages without a prefix that several adapters import.

| Package | Used by | Exports |
|---|---|---|
| `core/copilot/response` (package `copilotResponse`) | `openai`, `codex`, `copilot`, `core/xai` | `ConvertInput`, `ConvertTools`, `ConvertOutput`; Responses API wire types `Output`, `ToolCall` |
| `core/xai` | `grok`, `grok-oauth` | `ResponsesAPI` (`https://api.x.ai/v1/responses`), `BuildBody(ctx, model, messages, tools, effort string, fast bool) map[string]any`, `ParseSSEStream(label string, resp *http.Response) (*llmrouter.Output, error)` |

`xai.BuildBody` always requests a stream, so the non-streaming `Send` of both Grok adapters reads the SSE body with `xai.ParseSSEStream` and returns one assembled `Output`.

## Related

- [Agent Contract](/core-concepts): why capabilities are separate interfaces
