# go-llm-router - 架構

> 返回 [README](./README.zh.md)

## 目錄

- [概覽](#概覽)
- [模組：core](#模組core)
- [模組：router](#模組router)
- [模組：供應商轉接層](#模組供應商轉接層)
- [模組：串流](#模組串流)
- [模組：oauth](#模組oauth)
- [模組：多模態](#模組多模態)
- [模組：模型清單與篩選](#模組模型清單與篩選)
- [資料流](#資料流)
- [狀態機](#狀態機)

## 概覽

```mermaid
graph TB
    App[呼叫端] --> Router[router.New]
    Router --> Registry[newFn 前綴表]
    Registry --> KeyAgents[金鑰型 Agent]
    Registry --> OAuthAgents[OAuth 型 Agent]
    Registry --> CompatAgent[compat Agent]

    KeyAgents --> Core[core 契約層]
    OAuthAgents --> Core
    CompatAgent --> Core

    OAuthAgents --> OAuth[core/oauth]
    OAuth --> Keychain[go-pkg keychain]

    Core --> Stream[串流事件正規化]
    Core --> Usage[用量正規化]
    Core --> Media[圖片 / 語音介面]
    Core --> HTTP[go-pkg http]
```

`core` 只定義契約與共用行為，不認識任何具體供應商；供應商套件單向相依 `core`，彼此不互相引用（唯二例外：`grokOauth` 共用 `grok.RequestImage`、`openaiCodex` 共用 `openai` 的圖片 helper）。

## 模組：core

定義 Agent 契約、訊息與用量型別，以及所有供應商共用的推理／模式／串流／模型分類邏輯。

```mermaid
graph TB
    subgraph core
        Type[type.go<br/>Agent / Message / Output / Usage]
        Reasoning[reasoning.go<br/>Reasoning / ModelFilter]
        Mode[mode.go<br/>Mode / SupportFast]
        Provider[provider.go<br/>SupportTemperature / ResponsesAPI]
        Stream[stream.go<br/>OpenStream / StreamChat / StreamResponses]
        SSE[sse.go<br/>ScanSSE]
        Image[image.go<br/>ImageAgent / ImagePixelSize]
        Audio[audio.go<br/>STTAgent / TTSAgent / WrapPCM16]
    end
    Stream --> SSE
    Image --> Type
    Audio --> Type
    Reasoning --> Type
```

| 檔案 | 職責 |
|---|---|
| `type.go` | `Agent` / `StreamAgent` 介面、訊息與工具型別、`Usage.UnmarshalJSON` 的跨供應商欄位吸收 |
| `reasoning.go` | 六級推理等級、別名解析、`ClampReasoning`、模型標記分類與 `ModelFilter` |
| `mode.go` | `default` / `fast` 模式與各家 fast 層級的支援判斷 |
| `provider.go` | `temperature` 支援、Responses API 選路、共用 HTTP client |
| `stream.go` / `sse.go` | SSE 掃描、兩種串流格式的事件正規化、錯誤包裝與讀取上限 |
| `image.go` / `audio.go` | 多模態選用介面與尺寸／取樣率／WAV 封裝等純函式 |

## 模組：router

唯一的組裝點：把 `provider@model` 字串轉成具體 Agent。

```mermaid
graph TB
    Name["Name: provider@model"] --> Cut[切出前綴與模型]
    Cut --> Lookup{前綴在 newFn？}
    Lookup -- 是 --> Build[呼叫對應建構式]
    Lookup -- 否且含 @ --> Rewrite["改寫為 compat[prefix]@model"]
    Lookup -- 否且無 @ --> Err[unknown provider 錯誤]
    Rewrite --> Build
    Build --> Agent[core.Agent]
```

前綴另支援方括號實例名（`compat[lmstudio]@model`），`compatPrefix` 取出實例名後成為 Agent `Name()` 的前綴，讓同一個 compat 轉接層可同時代表多個端點。

## 模組：供應商轉接層

每個供應商套件的檔案佈局一致，差異只在轉換邏輯。

```mermaid
graph TB
    subgraph provider
        New[new.go<br/>Agent 結構 / Prefix / Name]
        Send[send.go<br/>訊息轉換 / 非串流請求]
        StreamFile[stream.go<br/>串流請求]
        Models[models.go<br/>模型清單 + 篩選]
        ReasoningFile[reasoning.go<br/>ReasoningLimits]
        Extra[usage.go / image.go / audio.go<br/>選用能力]
    end
    New --> Send
    New --> StreamFile
    Send --> CoreHTTP[core / go-pkg http]
    StreamFile --> CoreStream[core.OpenStream]
    Models --> Filter[core.MatchModelFilter]
```

| 群組 | 套件 | 線路格式 |
|---|---|---|
| OpenAI 相容 | `openai`、`deepseek`、`mistral`、`nvidia`、`openRouter`、`ollamaCloud`、`cloudflare`、`compat` | Chat Completions；OpenAI 新世代模型改走 Responses |
| Anthropic | `claude` | Messages API，thinking 預算換算 |
| Google | `gemini` | `:generateContent`，含 schema 淨化與 `cachedContents` |
| xAI | `grok`、`grokOauth` | Responses API + 圖片端點 |
| OAuth 代理 | `copilot`、`openaiCodex` | 廠商內部 Responses 端點，需 session 權杖 |

## 模組：串流

兩種上游格式共用同一組事件型別，呼叫端不需要分辨來源。

```mermaid
graph LR
    Req[OpenStream] --> Check{Content-Type}
    Check -- text/event-stream --> Scan[ScanSSE]
    Check -- 其他 --> Unsupported[StreamError + ErrStreamUnsupported]
    Scan --> Chat[StreamChat<br/>Chat Completions]
    Scan --> Responses[StreamResponses<br/>Responses API]
    Chat --> Events[StreamEvent channel]
    Responses --> Events
    Events --> Text[text]
    Events --> Reason[reasoning]
    Events --> Tool[tool_call]
    Events --> UsageEvt[usage]
    Events --> Done[done / error]
```

body 以 64 MiB `io.LimitReader` 封頂，錯誤訊息中的 frame 截至 512 bytes，錯誤 body 截至 8 KiB。

## 模組：oauth

三家 OAuth 供應商共用同一組流程骨架，權杖統一存放 keychain。

```mermaid
graph TB
    subgraph oauth
        Login[LoginWithCallback<br/>device flow / PKCE]
        Store[Load / HasToken / ClearToken]
        Fresh[EnsureFresh / EnsureFreshSession]
    end
    Login --> Keychain[(keychain)]
    Store --> Keychain
    Fresh --> Keychain
    Fresh --> Agent[Agent.authHeader]
    Agent --> Upstream[供應商端點]
```

`Load` 先讀新鍵，失敗時回退舊的 `agenvoy.*` 鍵；Copilot 多一層短期 session token 交換。

## 模組：多模態

```mermaid
graph TB
    ImageAgent[core.ImageAgent] --> OpenAIImg[openai<br/>/v1/images/*]
    ImageAgent --> CodexImg[codex<br/>Responses image_generation]
    ImageAgent --> GeminiImg[gemini<br/>:generateContent]
    ImageAgent --> GrokImg[grok / grok-oauth<br/>/v1/images/*]
    STT[core.STTAgent] --> OpenAISTT[openai<br/>/v1/audio/transcriptions]
    STT --> GeminiSTT[gemini<br/>逐字轉寫]
    TTS[core.TTSAgent] --> OpenAITTS[openai<br/>/v1/audio/speech]
    TTS --> GeminiTTS[gemini<br/>AUDIO 模態 + WrapPCM16]
```

圖片與語音模型都由 agent 模型決定，唯 `codex` 由後端以對話模型生成。

## 模組：模型清單與篩選

```mermaid
graph LR
    Models[provider.Models] --> Fetch[抓取供應商模型清單]
    Fetch --> Match[core.MatchModelFilter]
    Match --> TextOnly[TextOnly]
    Match --> STTOnly[STTOnly]
    Match --> TTSOnly[TTSOnly]
    Match --> ImageOnly[ImageOnly]
    ImageOnly --> NoVideo[排除 video 標記]
    Match --> IDs["[]string 模型 ID"]
    Fetch --> Infos[ModelInfos<br/>thinking / efforts / endpoints]
```

Cloudflare 例外：改以 Workers AI 的 task 名稱（`Text Generation` / `Automatic Speech Recognition` / `Text-to-Speech`）篩選，而非模型 ID 標記。

## 資料流

```mermaid
sequenceDiagram
    participant Caller as 呼叫端
    participant Router as router.New
    participant Agent as provider.Agent
    participant Core as core
    participant API as 供應商端點

    Caller->>Router: Config{Name, APIKey/Token}
    Router-->>Caller: core.Agent
    Caller->>Agent: Send / SendStream
    Agent->>Core: ClampReasoning / SupportFast / SupportTemperature
    Agent->>Agent: 訊息與工具轉換
    Agent->>API: HTTP 請求
    alt 串流
        API-->>Agent: SSE frames
        Agent->>Core: StreamChat / StreamResponses
        Core-->>Caller: StreamEvent channel
    else 非串流
        API-->>Agent: JSON
        Agent->>Core: Usage.UnmarshalJSON
        Agent-->>Caller: Output + status code
    end
```

## 狀態機

OAuth 權杖生命週期：

```mermaid
stateDiagram-v2
    [*] --> 未登入
    未登入 --> 已登入: LoginWithCallback
    已登入 --> 有效: EnsureFresh（未過期）
    已登入 --> 換新中: 距過期 < 60 秒
    換新中 --> 有效: refresh 成功
    換新中 --> 未登入: refresh 失敗
    有效 --> 已登入: 下次請求
    已登入 --> 未登入: ClearToken
```

串流事件生命週期：

```mermaid
stateDiagram-v2
    [*] --> 連線中
    連線中 --> 串流中: Content-Type 為 text/event-stream
    連線中 --> 失敗: 非 SSE 或上游錯誤
    串流中 --> 串流中: text / reasoning / tool_call / usage
    串流中 --> 完成: done
    串流中 --> 失敗: error frame
    完成 --> [*]
    失敗 --> [*]
```

***

©️ 2026 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
