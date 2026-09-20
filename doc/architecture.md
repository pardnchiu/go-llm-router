# go-llm-router - Architecture

> Back to [README](../README.md)

## Table of Contents

- [Overview](#overview)
- [Module: core](#module-core)
- [Module: router](#module-router)
- [Module: provider adapters](#module-provider-adapters)
- [Module: streaming](#module-streaming)
- [Module: oauth](#module-oauth)
- [Module: multimodal](#module-multimodal)
- [Module: model listing and filters](#module-model-listing-and-filters)
- [Data Flow](#data-flow)
- [State Machines](#state-machines)

## Overview

```mermaid
graph TB
    App[Caller] --> Router[router.New]
    Router --> Registry[newFn prefix table]
    Registry --> KeyAgents[Key-based agents]
    Registry --> OAuthAgents[OAuth agents]
    Registry --> CompatAgent[compat agent]

    KeyAgents --> Core[core contracts]
    OAuthAgents --> Core
    CompatAgent --> Core

    OAuthAgents --> OAuth[core/oauth]
    OAuth --> Keychain[go-pkg keychain]

    Core --> Stream[Stream event normalization]
    Core --> Usage[Usage normalization]
    Core --> Media[Image / audio interfaces]
    Core --> HTTP[go-pkg http]
```

`core` defines contracts and shared behavior and knows no concrete provider; provider packages depend on `core` one-way and never on each other, with exactly two exceptions: `grokOauth` reuses `grok.RequestImage`, and `openaiCodex` reuses the `openai` image helpers.

## Module: core

Holds the Agent contract, message and usage types, and the reasoning / mode / streaming / model-classification logic every provider shares.

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

| File | Responsibility |
|---|---|
| `type.go` | `Agent` / `StreamAgent` interfaces, message and tool types, cross-provider field absorption in `Usage.UnmarshalJSON` |
| `reasoning.go` | Six reasoning levels, alias parsing, `ClampReasoning`, marker-based model classification and `ModelFilter` |
| `mode.go` | `default` / `fast` modes and per-vendor fast-tier support |
| `provider.go` | `temperature` support, Responses API routing, the shared HTTP client |
| `stream.go` / `sse.go` | SSE scanning, event normalization for both stream formats, error wrapping and read caps |
| `image.go` / `audio.go` | Optional multimodal interfaces plus pure helpers for sizing, sample rate, and WAV framing |

## Module: router

The single assembly point: it turns a `provider@model` string into a concrete Agent.

```mermaid
graph TB
    Name["Name: provider@model"] --> Cut[Split prefix and model]
    Cut --> Lookup{Prefix in newFn?}
    Lookup -- yes --> Build[Call its constructor]
    Lookup -- "no, but has @" --> Rewrite["Rewrite to compat[prefix]@model"]
    Lookup -- "no, no @" --> Err[unknown provider error]
    Rewrite --> Build
    Build --> Agent[core.Agent]
```

Prefixes also carry a bracketed instance name (`compat[lmstudio]@model`); `compatPrefix` extracts it and it becomes the Agent's `Name()` prefix, so one compat adapter can represent several endpoints at once.

## Module: provider adapters

Every provider package uses the same file layout; only the conversion logic differs.

```mermaid
graph TB
    subgraph provider
        New[new.go<br/>Agent struct / Prefix / Name]
        Send[send.go<br/>Message conversion / non-stream request]
        StreamFile[stream.go<br/>Streaming request]
        Models[models.go<br/>Model listing + filter]
        ReasoningFile[reasoning.go<br/>ReasoningLimits]
        Extra[usage.go / image.go / audio.go<br/>Optional capabilities]
    end
    New --> Send
    New --> StreamFile
    Send --> CoreHTTP[core / go-pkg http]
    StreamFile --> CoreStream[core.OpenStream]
    Models --> Filter[core.MatchModelFilter]
```

| Group | Packages | Wire format |
|---|---|---|
| OpenAI-compatible | `openai`, `deepseek`, `mistral`, `nvidia`, `openRouter`, `ollamaCloud`, `cloudflare`, `compat` | Chat Completions; newer OpenAI generations switch to Responses |
| Anthropic | `claude` | Messages API with thinking-budget conversion |
| Google | `gemini` | `:generateContent` with schema sanitization and `cachedContents` |
| xAI | `grok`, `grokOauth` | Responses API plus the image endpoints |
| OAuth proxies | `copilot`, `openaiCodex` | Vendor-internal Responses endpoints requiring a session token |

## Module: streaming

Both upstream formats collapse onto one event type, so callers never tell them apart.

```mermaid
graph LR
    Req[OpenStream] --> Check{Content-Type}
    Check -- text/event-stream --> Scan[ScanSSE]
    Check -- anything else --> Unsupported[StreamError + ErrStreamUnsupported]
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

Bodies are capped by a 64 MiB `io.LimitReader`, frames quoted in errors at 512 bytes, and error bodies at 8 KiB.

## Module: oauth

All three OAuth providers share one flow skeleton and store tokens in the keychain.

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
    Agent --> Upstream[Provider endpoint]
```

`Load` reads the current key first and falls back to the legacy `agenvoy.*` key; Copilot adds one more exchange for a short-lived session token.

## Module: multimodal

```mermaid
graph TB
    ImageAgent[core.ImageAgent] --> OpenAIImg[openai<br/>/v1/images/*]
    ImageAgent --> CodexImg[codex<br/>Responses image_generation]
    ImageAgent --> GeminiImg[gemini<br/>:generateContent]
    ImageAgent --> GrokImg[grok / grok-oauth<br/>/v1/images/*]
    STT[core.STTAgent] --> OpenAISTT[openai<br/>/v1/audio/transcriptions]
    STT --> GeminiSTT[gemini<br/>verbatim transcript]
    TTS[core.TTSAgent] --> OpenAITTS[openai<br/>/v1/audio/speech]
    TTS --> GeminiTTS[gemini<br/>AUDIO modality + WrapPCM16]
```

Image and audio models are the agent model; `codex` alone generates from its chat model on the backend.

## Module: model listing and filters

```mermaid
graph LR
    Models[provider.Models] --> Fetch[Fetch the provider's model list]
    Fetch --> Match[core.MatchModelFilter]
    Match --> TextOnly[TextOnly]
    Match --> STTOnly[STTOnly]
    Match --> TTSOnly[TTSOnly]
    Match --> ImageOnly[ImageOnly]
    ImageOnly --> NoVideo[Excludes video markers]
    Match --> IDs["[]string of model IDs"]
    Fetch --> Infos[ModelInfos<br/>thinking / efforts / endpoints]
```

Cloudflare is the exception: it filters on the Workers AI task name (`Text Generation` / `Automatic Speech Recognition` / `Text-to-Speech`) rather than on model-ID markers.

## Data Flow

```mermaid
sequenceDiagram
    participant Caller
    participant Router as router.New
    participant Agent as provider.Agent
    participant Core as core
    participant API as Provider endpoint

    Caller->>Router: Config{Name, APIKey/Token}
    Router-->>Caller: core.Agent
    Caller->>Agent: Send / SendStream
    Agent->>Core: ClampReasoning / SupportFast / SupportTemperature
    Agent->>Agent: Convert messages and tools
    Agent->>API: HTTP request
    alt Streaming
        API-->>Agent: SSE frames
        Agent->>Core: StreamChat / StreamResponses
        Core-->>Caller: StreamEvent channel
    else Non-streaming
        API-->>Agent: JSON
        Agent->>Core: Usage.UnmarshalJSON
        Agent-->>Caller: Output + status code
    end
```

## State Machines

OAuth token lifecycle:

```mermaid
stateDiagram-v2
    [*] --> LoggedOut
    LoggedOut --> LoggedIn: LoginWithCallback
    LoggedIn --> Valid: EnsureFresh (not expired)
    LoggedIn --> Refreshing: less than 60s to expiry
    Refreshing --> Valid: refresh succeeded
    Refreshing --> LoggedOut: refresh failed
    Valid --> LoggedIn: next request
    LoggedIn --> LoggedOut: ClearToken
```

Stream event lifecycle:

```mermaid
stateDiagram-v2
    [*] --> Connecting
    Connecting --> Streaming: Content-Type is text/event-stream
    Connecting --> Failed: non-SSE or upstream error
    Streaming --> Streaming: text / reasoning / tool_call / usage
    Streaming --> Completed: done
    Streaming --> Failed: error frame
    Completed --> [*]
    Failed --> [*]
```

***

©️ 2026 [邱敬幃 Pardn Chiu](https://www.linkedin.com/in/pardnchiu)
