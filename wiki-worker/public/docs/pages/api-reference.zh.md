# 介面與工廠

Agent 介面、`router` 工廠，以及各供應商套件的匯出。所有型別位於 `github.com/pardnchiu/go-llm-router/core`（該目錄宣告的 package 名為 `llmrouter`），工廠位於 `core/router`。

## 介面

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

| 符號 | 簽名 | 說明 |
|---|---|---|
| `router.Config` | `struct{ Name, APIKey string; Token any; BaseURL, AccountID, GatewayID string; EnableClaude bool; StateDir string }` | `Name` 為 `provider@model`；`EnableClaude`／`StateDir` 僅用於 `claude-code@` |
| `router.New` | `func(config Config) (llmrouter.Agent, error)` | 依前綴建構 Agent；未知前綴且含 `@` 時改寫為 `compat[<prefix>]@<model>` |

## 供應商名稱

`core` 為每個供應商前綴宣告一個常數，不含結尾的 `@`。以 `llmrouter.PROVIDER_GEMINI + "@" + model` 組出名稱；`Providers()` 回傳 `compat` 以外的所有前綴。

| 常數 | 值 |
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

## 供應商套件

每個供應商套件皆提供 `New(llmrouter.Config) (*Agent, error)`、`(*Agent).Name()` 與 `Models(ctx, llmrouter.Config, llmrouter.ModelFilter) ([]string, error)`。`core/compat` 沒有 `Models`，`core/claudeCode.Models(ctx, llmrouter.Config)` 不接受篩選條件。

| 套件 | 前綴 | 額外匯出 |
|---|---|---|
| `core/claude` | `claude@` | Messages API 傳輸型別 `Output`、`Content` |
| `core/claudeCode` | `claude-code@` | `Usage`、`CheckBinary`、`Is` |
| `core/openai` | `openai@` | `ImageInput`、`ReadImageStream`、`ImageInstructions` |
| `core/gemini` | `gemini@` | `ModelInfos`；`:generateContent` 傳輸型別 `Content`、`Part`、`InlineData`、`FunctionCall`、`FunctionResponse`、`Output` |
| `core/grok` | `grok@` | `RequestImage` |
| `core/grokOauth` | `grok-oauth@` | `Usage` |
| `core/deepseek` | `deepseek@` | `Usage` |
| `core/mistral` | `mistral@` | `ModelInfos` |
| `core/nvidia` | `nvidia@` | — |
| `core/ollamaCloud` | `ollama-cloud@` | — |
| `core/openRouter` | `openrouter@` | `Usage` |
| `core/cloudflare` | `cloudflare@` | — |
| `core/compat` | `compat@` / `compat[name]@` | — |
| `core/copilot` | `copilot@` | `Usage`、`ModelInfos` |
| `core/openaiCodex` | `codex@` | `Usage` |

## 共用傳輸輔助

沒有前綴、由多個轉接層匯入的套件。

| 套件 | 使用者 | 匯出 |
|---|---|---|
| `core/copilot/response`（package `copilotResponse`） | `openai`、`codex`、`copilot`、`core/xai` | `ConvertInput`、`ConvertTools`、`ConvertOutput`；Responses API 傳輸型別 `Output`、`ToolCall` |
| `core/xai` | `grok`、`grok-oauth` | `ResponsesAPI`（`https://api.x.ai/v1/responses`）、`BuildBody(ctx, model, messages, tools, effort string, fast bool) map[string]any`、`ParseSSEStream(label string, resp *http.Response) (*llmrouter.Output, error)` |

`xai.BuildBody` 一律要求串流，因此兩個 Grok 轉接層的非串流 `Send` 也是以 `xai.ParseSSEStream` 讀取 SSE，再組成單一 `Output` 回傳。

## 相關頁面

- [Agent 契約](/zh/core-concepts)：能力為何拆成獨立介面
