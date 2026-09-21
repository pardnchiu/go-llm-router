# go-llm-router - 技術文件

> 返回 [README](./README.zh.md)

## 目錄

- [前置需求](#前置需求)
- [安裝](#安裝)
- [設定](#設定)
- [使用方式](#使用方式)
- [API 參考](#api-參考)

## 前置需求

- Go 1.25 或以上
- `github.com/pardnchiu/go-pkg` v0.13.14（唯一直接相依，`go get` 時自動取得）
- 至少一組供應商憑證：
  - 金鑰型：Anthropic、OpenAI、Gemini、xAI、DeepSeek、Mistral、NVIDIA、Ollama Cloud、OpenRouter、Cloudflare Workers AI
  - OAuth 型：GitHub Copilot 訂閱、ChatGPT（Codex）帳號、Grok 帳號
  - 自架或第三方 OpenAI 相容端點（`compat`）
- macOS Keychain 或 Linux `secret-tool`（僅 OAuth 供應商需要，權杖由 `go-pkg/filesystem/keychain` 保存）

## 安裝

### 以模組相依安裝

```bash
go get github.com/pardnchiu/go-llm-router
```

### 從原始碼取得

```bash
git clone https://github.com/pardnchiu/go-llm-router.git
cd go-llm-router
go build ./...
```

## 設定

### 函式庫本身

函式庫不讀取任何環境變數，憑證一律由呼叫端透過 `router.Config` 傳入。OAuth 供應商的權杖則存於系統 keychain：

| 供應商 | Keychain 鍵 | 取得方式 |
|---|---|---|
| Copilot | `COPILOT_OAUTH_TOKEN` | `core/oauth/copilot.LoginWithCallback`（device flow） |
| Codex | `CODEX_OAUTH_TOKEN` | `core/oauth/codex.LoginWithCallback`（PKCE 瀏覽器授權） |
| Grok | `GROK_OAUTH_TOKEN` | `core/oauth/grok.LoginWithCallback`（PKCE 瀏覽器授權） |

三者皆相容舊鍵 `agenvoy.<provider>.token`，讀取時自動回退。

## 使用方式

### 基礎：取得 Agent 並送出請求

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

### 串流

`SendStream` 是選用介面，以型別斷言取得；不支援串流的 Agent 會斷言失敗。

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

### 推理等級與 fast 模式

```go
reasoning, ok := llmrouter.ParseReasoning("xhigh") // none / low / medium / high / xhigh / max
if !ok {
	reasoning = llmrouter.ReasoningDefault          // medium
}

// 查詢模型實際支援的區間，超出範圍的請求會被自動收斂
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

### 工具呼叫

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

工具結果以 `Role: "tool"` + `ToolCallID` 回傳；Gemini 的 `ThoughtSignature` 需原樣帶回，否則後續回合會被拒絕。

### 多模態輸入

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

### 列出模型

```go
ids, err := gemini.Models(ctx, llmrouter.Config{APIKey: key}, llmrouter.ModelFilter{TextOnly: true})
if err != nil {
	return err
}

// 只要圖片模型；video 模型不會混入
images, err := gemini.Models(ctx, llmrouter.Config{APIKey: key}, llmrouter.ModelFilter{ImageOnly: true})

// gemini / mistral / copilot 另有 ModelInfos，回傳 thinking、efforts、endpoints
infos, err := gemini.ModelInfos(ctx, llmrouter.Config{APIKey: key}, llmrouter.ModelFilter{TextOnly: true})
```

### 圖片生成

圖片模型即 agent 模型，與對話、TTS 的選法一致；`codex` 例外，後端以對話模型直接生成。

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

### 語音轉文字與文字轉語音

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

### 查詢餘額與配額

```go
remaining, err := openrouter.Usage(ctx, llmrouter.Config{APIKey: key})
if err != nil {
	return err
}
fmt.Printf("credits left: %.2f\n", remaining)
```

### 自架或第三方相容端點

```go
// 明確指定 compat
agent, err := router.New(router.Config{
	Name:    "compat[lmstudio]@qwen3-30b",
	BaseURL: "http://127.0.0.1:1234/v1",
	APIKey:  "not-needed",
})

// 未知前綴自動落到 compat，Name() 會回 "groq@llama-3.3-70b"
agent, err = router.New(router.Config{
	Name:    "groq@llama-3.3-70b",
	BaseURL: "https://api.groq.com/openai/v1",
	APIKey:  os.Getenv("GROQ_API_KEY"),
})
```

`BaseURL` 留空時預設為 `http://localhost:11434/v1`（Ollama）。

## API 參考

### 介面

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

`Agent` 為必要實作，其餘皆以型別斷言取得。

import 路徑為 `github.com/pardnchiu/go-llm-router/core`，該目錄宣告的 package 名為 `llmrouter`，因此符號寫作 `llmrouter.Agent`。

### router

| 符號 | 簽名 | 說明 |
|---|---|---|
| `router.Config` | `struct{ Name, APIKey string; Token any; BaseURL, AccountID, GatewayID string }` | `Name` 為 `provider@model`；`Token` 供 OAuth 供應商傳入 `*llmrouter.CopilotToken` / `*llmrouter.CodexToken` / `*llmrouter.GrokToken` |
| `router.New` | `func(config Config) (llmrouter.Agent, error)` | 依前綴建立 Agent；未知前綴且含 `@` 時改建 `compat[<prefix>]@<model>` |

### 供應商前綴

| 前綴 | 套件 | 憑證 | 串流 | 額外能力 |
|---|---|---|---|---|
| `claude@` | `core/claude` | `APIKey` | ✓ | — |
| `openai@` | `core/openai` | `APIKey` | ✓ | 圖片、STT、TTS |
| `gemini@` | `core/gemini` | `APIKey` | ✓ | 圖片、STT、TTS、`cachedContents` 前綴快取 |
| `grok@` | `core/grok` | `APIKey` | ✓ | 圖片 |
| `deepseek@` | `core/deepseek` | `APIKey` | ✓ | `Usage` |
| `mistral@` | `core/mistral` | `APIKey` | ✓ | `ModelInfos` |
| `nvidia@` | `core/nvidia` | `APIKey` | ✓ | — |
| `ollama-cloud@` | `core/ollamaCloud` | `APIKey` | ✓ | `Usage` |
| `openrouter@` | `core/openRouter` | `APIKey` | ✓ | 圖片、STT、TTS、`Usage` |
| `cloudflare@` | `core/cloudflare` | `APIKey` + `AccountID` + `GatewayID` | ✓ | — |
| `compat@` / `compat[name]@` | `core/compat` | `APIKey` + `BaseURL` | ✓ | — |
| `copilot@` | `core/copilot` | `*llmrouter.CopilotToken` | ✓ | `Usage`、`ModelInfos` |
| `codex@` | `core/openaiCodex` | `*llmrouter.CodexToken` | ✓ | 圖片、`Usage` |
| `grok-oauth@` | `core/grokOauth` | `*llmrouter.GrokToken` | ✓ | 圖片、`Usage` |

每個供應商套件都提供 `New(llmrouter.Config) (*Agent, error)` 與 `Models(ctx, llmrouter.Config, llmrouter.ModelFilter) ([]string, error)`。

### 推理等級

| 常數 | 字串 | 別名 |
|---|---|---|
| `ReasoningNone` | `none` | — |
| `ReasoningLow` | `low` | `minimal` |
| `ReasoningMedium` | `medium`（`ReasoningDefault`） | — |
| `ReasoningHigh` | `high` | — |
| `ReasoningXHigh` | `xhigh` | `extra` |
| `ReasoningMax` | `max` | `ultra` |

| 函式 | 簽名 | 說明 |
|---|---|---|
| `ParseReasoning` | `func(s string) (Reasoning, bool)` | 無法解析時回 `ReasoningDefault, false` |
| `ClampReasoning` | `func(r, lo, hi Reasoning, provider, model string) Reasoning` | 收斂時輸出 debug log |
| `OpenAIEffortRange` | `func(model string) (low, high Reasoning)` | 依 OpenAI 模型世代決定 effort 區間 |

### 模式

| 符號 | 簽名 | 說明 |
|---|---|---|
| `ModeDefault` / `ModeFast` | `Mode` | `fast` 對映各家的優先處理層級 |
| `ParseMode` | `func(s string) (Mode, bool)` | 接受 `default` / `fast` |
| `SupportFast` | `func(providerName, model string) bool` | 依模型世代判斷是否具備 fast 層級 |
| `WarnFastDowngrade` | `func(providerName, model, tier string)` | 回應層級非 `fast` / `priority` 時發出警告 |

### 模型篩選

| 符號 | 簽名 | 說明 |
|---|---|---|
| `ModelFilter` | `struct{ TextOnly, STTOnly, TTSOnly, ImageOnly bool }` | 多個條件同時成立時取交集 |
| `IsTextModel` / `IsSTTModel` / `IsTTSModel` / `IsImageModel` | `func(id string) bool` | 依模型 ID 標記判斷；`IsImageModel` 排除 video 模型 |
| `MatchModelFilter` | `func(id string, filter ModelFilter) bool` | `openai`、`gemini`、`grok`、`grok-oauth` 的清單使用 |
| `ModelInfo` | `struct{ ID string; Thinking bool; Efforts, Endpoints []string }` | `ModelInfos` 的回傳元素 |

各供應商支援的旗標不同，不支援的旗標會被忽略而非報錯：

| 供應商 | `TextOnly` | `STTOnly` | `TTSOnly` | `ImageOnly` | 來源 |
|---|---|---|---|---|---|
| `openai`、`gemini`、`grok`、`grok-oauth` | ✓ | ✓ | ✓ | ✓ | 模型 ID 標記 |
| `openrouter` | ✓ | ✓ | ✓ | ✓ | `/api/v1/images/models`、`/api/v1/models?output_modalities=speech` / `transcription` |
| `cloudflare` | ✓ | ✓ | ✓ | — | Workers AI task 名稱 |
| `claude`、`copilot`、`deepseek`、`mistral`、`nvidia`、`ollama-cloud`、`codex` | ✓ | — | — | — | 模型 ID 標記 |

### 訊息與輸出

| 型別 | 說明 |
|---|---|
| `Message` | `Role` / `Content`（`string` 或 `[]ContentPart`）/ `ReasoningContent` / `ToolCalls` / `ToolCallID` |
| `ContentPart`、`ImageURL` | `text` 與 `image_url` 內容部分，`image_url` 接受 data URI |
| `Tool`、`ToolFunction` | OpenAI 形狀的工具定義，`Parameters` 為 `json.RawMessage` |
| `ToolCall` | 含 Gemini 專用的 `ThoughtSignature` |
| `Output`、`OutputChoices` | `Choices` / `Usage` / `ServiceTier` / `Error` |
| `Usage` | `Input` / `Output` / `CacheCreate` / `CacheRead`，自訂 `UnmarshalJSON` 吸收各家欄位命名 |

`Usage.UnmarshalJSON` 會把 `prompt_tokens` 與 `input_tokens`、`completion_tokens` 與 `output_tokens` 相加，並把 `prompt_tokens_details.cached_tokens` 自 `Input` 扣除後計入 `CacheRead`，避免快取命中被重複計價。

### 串流

| 符號 | 說明 |
|---|---|
| `StreamEvent` | `Type` 為 `text` / `reasoning` / `tool_call` / `usage` / `done` / `error` |
| `ToolCallDelta` | 串流中的工具呼叫增量，含 `Index` 與 `ThoughtSignature` |
| `StreamError` | 帶 `Provider` / `Code` / `Body`，可 `errors.As` 取出上游狀態碼 |
| `ErrStreamUnsupported` | 上游回傳非 SSE 內容時包在 `StreamError.Err` 內 |
| `OpenStream` | `func(ctx, client, url, headers, body, label) (*http.Response, error)` |
| `StreamChat` / `StreamResponses` | 分別解析 Chat Completions 與 Responses API 的 SSE，回傳事件 channel |
| `ScanSSE` | `func(reader *bufio.Reader, handle func(event, data string) bool) error` |
| `StreamBodyLimit` / `ErrorBodyLimit` / `ErrorFrameLimit` / `JSONBodyLimit` | 64 MiB / 8 KiB / 512 B / 64 KiB 的讀取上限 |

### 圖片

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

| 供應商 | 圖片模型 | 端點 | 結果 |
|---|---|---|---|
| `openai` | agent 模型，如 `openai@gpt-image-2` | `/v1/images/generations`，帶參考圖時 `/v1/images/edits` | PNG（`output_format`），含 `Revised` |
| `codex` | 後端預設 | ChatGPT Codex Responses 的 `image_generation` 工具 | PNG，含 `Revised` |
| `gemini` | agent 模型，如 `gemini@gemini-3.1-flash-image` | `:generateContent` | JPEG |
| `grok` / `grok-oauth` | agent 模型，如 `grok@grok-imagine-image-2.0` | `/v1/images/generations` 與 `/v1/images/edits` | JPEG |
| `openrouter` | agent 模型，如 `openrouter@google/gemini-2.5-flash-image` | `/api/v1/images` | 依回應的 `media_type` |

| 選項 | `openai` | `codex` | `gemini` | `grok` / `grok-oauth` | `openrouter` |
|---|---|---|---|---|---|
| `AspectRatio` + `Size` | `size` 的 `WIDTHxHEIGHT` | 端點忽略 | `imageConfig.aspectRatio` / `.imageSize` | `aspect_ratio` / `resolution`，`4k` 收斂為 `2k` | `aspect_ratio` / `resolution` |
| `Quality` | `quality` | 端點忽略 | 無對應 | `quality` | `quality` |
| `RefImageB64` | `/v1/images/edits` 的 `image` 檔案欄位 | `input_image` 內容部分 | `inline_data` 部分 | `/v1/images/edits` 的 `image.url` | `input_references` 的 data URL |

`ImagePixelSize` 以 `Size` 設定**短邊**：`16:9` + `1k` 得到 `1824x1024`；改放長邊會低於端點的最小像素預算而被拒絕。ChatGPT Codex 後端固定回 `1254x1254` 且 `quality: low`，因此 `codex` 兩項都不送。

### 語音

```go
type STTOptions struct{ Prompt, Language, MimeType string }
type TTSOptions struct{ Voice, Format string } // wav（預設）/ mp3 / opus / aac / flac / pcm
```

| 供應商 | STT | TTS | 備註 |
|---|---|---|---|
| `openai` | `/v1/audio/transcriptions` | `/v1/audio/speech` | 模型即 agent 模型，預設聲音 `alloy` |
| `openrouter` | multipart `/api/v1/audio/transcriptions` | `/api/v1/audio/speech` | `Voice` 留空時取 OpenRouter 目錄中該模型 `supported_voices` 的第一個；不在清單內的 voice 直接在本地回錯並列出合法清單；只有 `mp3` 直接回傳，其餘 `Format` 一律請求 `pcm` 並經 `WrapPCM16` 輸出 WAV |
| `gemini` | `:generateContent` 逐字轉寫 | `:generateContent` 的 `AUDIO` 模態 | 預設聲音 `Kore`，PCM 由 `WrapPCM16` 封成 WAV |

| 函式 | 簽名 | 說明 |
|---|---|---|
| `PCMRate` | `func(mime string) int` | 自 `audio/L16;rate=24000` 取樣率，取不到時回 `24000` |
| `WrapPCM16` | `func(pcm []byte, rate int) []byte` | 補上 44 bytes 的 WAV header |
| `DataURI` | `func(mime, b64 string) string` | mime 為空時預設 `image/png` |

### OAuth

`core/oauth/copilot`、`core/oauth/codex`、`core/oauth/grok` 三個套件提供相同形狀的 API：

| 函式 | 簽名 | 說明 |
|---|---|---|
| `Load` | `func() (*llmrouter.XxxToken, error)` | 自 keychain 讀取並解碼權杖 |
| `HasToken` | `func() bool` | 判斷是否已登入 |
| `ClearToken` | `func() error` | 移除權杖，含舊鍵 |
| `LoginWithCallback` | `func(ctx, onCode/onURL) (*llmrouter.XxxToken, error)` | Copilot 為 device flow（回傳 `*DeviceCode`），Codex 與 Grok 為 PKCE 授權 URL |
| `EnsureFresh` / `EnsureFreshSession` | `func(ctx, token[, refresh]) (...)` | 過期前 60 秒換新；Copilot 另外換取短期 session token |

### 用量查詢

| 套件 | 簽名 | 回傳語意 |
|---|---|---|
| `core/openRouter` | `Usage(ctx, llmrouter.Config) (float64, error)` | 剩餘點數（`total_credits - total_usage`） |
| `core/deepseek` | 同上 | 帳戶餘額 |
| `core/ollamaCloud` | 同上 | 月配額剩餘百分比 |
| `core/copilot` | 同上（需 `Token`） | Chat 或 premium interactions 的剩餘百分比 |
| `core/openaiCodex` | 同上（需 `APIKey`、`AccountID`） | 主／次視窗中較緊的剩餘百分比 |
| `core/grokOauth` | 同上 | 剩餘點數百分比 |

### 其他

| 符號 | 簽名 | 說明 |
|---|---|---|
| `NewHTTPClient` | `func() *http.Client` | 10 分鐘 timeout；`codex` 與 `grok-oauth` 另外自建 15 秒 response header timeout 的 client |
| `SupportTemperature` | `func(providerName, model string) bool` | Claude、`gpt-5` 系列、`deepseek-reasoner`、Gemini preview 皆不接受 `temperature` |
| `ResponsesAPI` | `func(providerName, model string) bool` | 決定 OpenAI / Copilot 走 Responses 還是 Chat Completions |
| `TruncateFrame` | `func(data string) string` | 錯誤訊息中的 SSE frame 截至 512 bytes |
| `summary.VoiceReply` | `func(ctx, text string) (string, error)` | 超過 320 字元時以 Gemini 壓縮為可朗讀版本，金鑰取自 keychain 的 `GEMINI_API_KEY` |

***

©️ 2026 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
