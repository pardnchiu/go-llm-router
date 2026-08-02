# go-llm-router - Architecture

> Back to [README](../README.md)

## Overview

`go-llm-router` exposes a stable Go API over multiple LLM provider protocols. Callers either create a concrete provider directly or pass a `provider@model` identifier to `core/router`. Both paths use the same message, tool, reasoning, mode, response, and streaming contracts.

```mermaid
graph TB
    App[Caller Application]
    Server[cmd/test]
    Router[core/router]
    Contract[core Agent / StreamAgent]
    Shared[Shared Core Policies]
    Providers[Provider Agents]
    APIs[Provider APIs]
    OAuth[OAuth Packages]
    Keychain[System Keychain]

    App --> Router
    App --> Contract
    Server --> Router
    Router --> Contract
    Contract --> Shared
    Contract --> Providers
    Providers --> APIs
    OAuth --> Keychain
    OAuth --> Providers
```

| Layer | Packages | Responsibility |
|---|---|---|
| Caller | Application code, `cmd/test` | Build messages and tools; choose model, reasoning, mode, and synchronous or streaming consumption |
| Router | `core/router` | Parse a provider name and construct the matching `core.Agent` |
| Shared core | `core` | Define transport-neutral types, reasoning and fast-mode policy, and HTTP client defaults |
| Provider adapter | `core/<provider>` | Authenticate, convert payloads, select provider endpoints, and normalize responses |
| OAuth | `core/oauth/*` | Login, token storage, expiry checks, and refresh paths |

## Module: Core Contracts

`core/type.go` is the public cross-provider boundary. Adapters return normalized `Output`, `Usage`, and `StreamEvent` values rather than native upstream responses.

```mermaid
classDiagram
    class Agent {
        <<interface>>
        +Name() string
        +Send(ctx, messages, toolDefs, reasoning, mode) Output, int, error
    }
    class StreamAgent {
        <<interface>>
        +SendStream(ctx, messages, toolDefs, reasoning, mode) channel StreamEvent, error
    }
    class Message {
        +string Role
        +any Content
        +string ReasoningContent
        +ToolCall[] ToolCalls
        +string ToolCallID
    }
    class Tool {
        +string Type
        +ToolFunction Function
    }
    class Output {
        +OutputChoices[] Choices
        +Usage Usage
        +string ServiceTier
    }
    class StreamEvent {
        +StreamEventType Type
        +string TextDelta
        +string ReasoningDelta
        +ToolCallDelta ToolCall
        +Usage Usage
        +string FinishReason
        +error Err
    }

    Agent --> Message
    Agent --> Tool
    Agent --> Output
    StreamAgent --> StreamEvent
    Output --> Usage
```

| Type | Role |
|---|---|
| `Message` | A text or multipart conversation item, including reasoning, tool calls, and tool-result linkage |
| `Tool` / `ToolFunction` | OpenAI-style function definition with JSON Schema parameters |
| `Output` / `OutputChoices` | Normalized non-streaming response, finish reason, usage, optional service tier, and error payload |
| `Usage` | Unified input, output, cache-creation, and cache-read token counters |
| `StreamEvent` | Typed text, reasoning, tool-call, usage, completion, and error deltas |
| `Config` | Shared model, credential, custom endpoint, and Cloudflare account/gateway construction settings |

## Module: Router

`router.New` gets the substring before `@`, removes any `[tag]`, looks up the provider factory, and returns a unified Agent. Unknown provider keys return an error.

```mermaid
flowchart LR
    Name[router.Config.Name]
    Prefix[Split at @]
    Tag[Drop Optional Bracket Tag]
    Lookup{newFn entry?}
    Error[unknown provider error]
    Factory[Provider New]
    Agent[core.Agent]

    Name --> Prefix --> Tag --> Lookup
    Lookup -->|no| Error
    Lookup -->|yes| Factory --> Agent
```

| Provider key | Constructor | Required configuration |
|---|---|---|
| `claude`, `openai`, `gemini`, `grok`, `deepseek`, `nvidia`, `openrouter` | Matching package `New` | `APIKey` |
| `cloudflare` | `cloudflare.New` | `APIKey`, `AccountID`, `GatewayID` |
| `compat` | `compat.New` | `BaseURL`; optional `APIKey` |
| `copilot`, `codex`, `grok-oauth` | OAuth-backed package `New` | `Token` |

## Module: Request Policy

The core standardizes call parameters but deliberately leaves wire payloads to adapters.

```mermaid
flowchart TB
    Call[Send or SendStream]
    Reasoning[Reasoning]
    Mode[Mode]
    Capability[Provider / Model Capability]
    Shape[Provider Payload Conversion]
    Transport[JSON HTTP or SSE]
    Normalize[Output or StreamEvent]

    Call --> Reasoning
    Call --> Mode
    Reasoning --> Capability --> Shape
    Mode --> Capability
    Shape --> Transport --> Normalize
```

### Reasoning

`Reasoning` supports `none`, `low`, `medium`, `high`, `xhigh`, and `max`; `medium` is the default. `ParseReasoning` also accepts `minimal`, `extra`, and `ultra`. Provider adapters use range and effort helpers, including `ClampReasoning` and `OpenAIEffortRange`, before formatting native controls.

### Fast Mode

`ModeFast` is a capability request, not a universal guarantee. `core.SupportFast` checks the provider/model pair, while `core.WarnFastDowngrade` logs any returned downgrade.

| Provider route | Native fast-mode control |
|---|---|
| Claude | `speed: "fast"` and the fast-mode beta header |
| OpenAI | `service_tier: "fast"` |
| Grok and Grok OAuth | `service_tier: "priority"` |
| OpenRouter | `service_tier: "priority"` for supported routed models |
| Other adapters | No fast-tier field currently added |

## Module: Provider Adapters

Each adapter owns system-prompt handling, message and tool conversion, authentication, endpoint selection, and upstream response decoding.

```mermaid
graph TB
    Core[core.Message / core.Tool]
    Convert[Provider Converter]
    Chat[Chat Completions APIs]
    Responses[Responses APIs]
    Native[Native Provider APIs]
    Upstream[Response / SSE]
    Result[core.Output / StreamEvent]

    Core --> Convert
    Convert --> Chat
    Convert --> Responses
    Convert --> Native
    Chat --> Upstream
    Responses --> Upstream
    Native --> Upstream
    Upstream --> Result
```

| Module | API form | Implementation focus |
|---|---|---|
| `claude` | Anthropic Messages | Tool/image conversion, prompt caching, thinking, fast mode, and public streaming |
| `openai` | Chat Completions or Responses | Model-driven endpoint choice, `instructions`, reasoning effort, and fast tier |
| `gemini` | `generateContent` | Content and function-declaration conversion, cache support, schema cleanup, and synthetic-skill rewrite |
| `grok` | Chat Completions | Reasoning effort and priority service tier |
| `grokOauth` | Responses via SSE | OAuth authentication and internal SSE assembly into one output |
| `copilot` | Chat Completions or Responses | Endpoint-capability lookup/cache, OAuth authentication, and public streaming |
| `openaiCodex` | ChatGPT Codex Responses via SSE | OAuth authentication, prompt-cache key, internal SSE assembly, and image generation |
| `deepseek` | Chat Completions | System-prompt merge and assistant reasoning placeholder |
| `nvidia` | Chat Completions | System-prompt merge and reasoning-effort mapping |
| `openRouter` | Chat Completions | Reasoning-detail preservation and supported priority tier |
| `cloudflare` | Workers AI Run | Content flattening and account/gateway request configuration |
| `compat` | Custom Chat Completions | Standard OpenAI-style payload to a supplied base URL |

## Module: Public Streaming

Providers that implement `core.StreamAgent` translate upstream SSE events to common event values. The caller can use a type assertion without depending on provider-specific event names.

```mermaid
sequenceDiagram
    participant Client as Caller
    participant Agent as StreamAgent
    participant API as Provider SSE API

    Client->>Agent: SendStream(ctx, messages, tools, reasoning, mode)
    Agent->>API: POST stream=true
    API-->>Agent: text delta
    Agent-->>Client: StreamEventText
    API-->>Agent: reasoning delta
    Agent-->>Client: StreamEventReasoning
    API-->>Agent: tool-call delta
    Agent-->>Client: StreamEventToolCall
    API-->>Agent: usage and completion
    Agent-->>Client: StreamEventUsage and StreamEventDone
```

| Event | Meaning |
|---|---|
| `text` | Output-text delta |
| `reasoning` | Reasoning-summary or reasoning-text delta |
| `tool_call` | Tool-call ID, name, or argument fragment |
| `usage` | Normalized token counters |
| `done` | Completion reason such as `stop` or `tool_calls` |
| `error` | Upstream, read, or decode error |

Claude and Copilot expose `SendStream`. Codex and Grok OAuth consume provider SSE inside their public `Send` implementation, then return a completed `core.Output`.

## Module: OAuth Lifecycle

OAuth adapters persist JSON token values in the operating-system keychain. Codex and Grok expiry checks use a 60-second safety buffer, then refresh and persist a replacement token when needed.

```mermaid
stateDiagram-v2
    [*] --> Missing: no stored token
    Missing --> Authorizing: LoginWithCallback
    Authorizing --> Active: exchange succeeds
    Active --> Active: Load valid token
    Active --> Expired: expiry safety window
    Expired --> Refreshing: EnsureFresh
    Refreshing --> Active: save refreshed token
    Refreshing --> Missing: refresh fails or ClearToken
    Active --> Missing: ClearToken
```

```mermaid
graph LR
    Login[LoginWithCallback]
    Exchange[Authorization-code Exchange]
    Token[core Token]
    Store[System Keychain]
    Load[Load / HasToken]
    Fresh[EnsureFresh]
    Refresh[Refresh Grant]
    Agent[OAuth Provider Agent]

    Login --> Exchange --> Token --> Store
    Store --> Load --> Fresh
    Fresh -->|valid| Agent
    Fresh -->|expired| Refresh --> Token
```

| OAuth package | Token type | Primary key |
|---|---|---|
| `core/oauth/copilot` | `core.CopilotToken` | `COPILOT_OAUTH_TOKEN`, plus legacy compatibility |
| `core/oauth/codex` | `core.CodexToken` | `CODEX_OAUTH_TOKEN`, plus `agenvoy.codex.token` compatibility |
| `core/oauth/grok` | `core.GrokToken` | `GROK_OAUTH_TOKEN`, plus `agenvoy.grok-oauth.token` compatibility |

## Module: HTTP-Compatible Test Server

`cmd/test` is a manual OpenAI-compatible wrapper, not a production gateway. It validates the request, resolves credentials from environment variables, constructs a router Agent, and returns JSON or OpenAI-style SSE chunks.

```mermaid
sequenceDiagram
    participant Client as HTTP Client
    participant Server as cmd/test
    participant Config as resolveConfig
    participant Router as router.New
    participant Agent as Agent or StreamAgent
    participant API as Provider API

    Client->>Server: POST /v1/chat/completions
    Server->>Server: validate model, reasoning, mode
    Server->>Config: read provider environment
    Config-->>Server: router.Config
    Server->>Router: construct Agent
    Router-->>Server: core.Agent
    alt stream is false
        Server->>Agent: Send
        Agent->>API: JSON request
        API-->>Agent: response
        Agent-->>Server: core.Output
        Server-->>Client: JSON
    else stream is true
        Server->>Agent: SendStream
        Agent->>API: SSE request
        API-->>Agent: events
        Server-->>Client: SSE chunks and [DONE]
    end
```

| Item | Value |
|---|---|
| Default port | `8787`; override with `PORT` |
| Endpoint | `POST /v1/chat/completions` |
| Non-streaming | `Agent.Send` returns JSON `core.Output` |
| Streaming | `StreamAgent.SendStream` is translated to `text/event-stream` |
| `make test` | `go run ./cmd/test` |
| `make build` | `go build ./...` |
| `make vet` | `go vet ./...` |

## End-to-End Data Flow

```mermaid
flowchart TB
    Input[Messages and Tools]
    Config[router.Config]
    Route[router.New]
    Policy[Reasoning and Mode]
    Agent[Provider Agent]
    Shape[Message Tool Credential Conversion]
    Transport[JSON or SSE HTTP]
    Remote[Provider Model]
    Normalize[Output or StreamEvent]
    Consumer[Caller]

    Input --> Config --> Route --> Agent
    Policy --> Agent
    Agent --> Shape --> Transport --> Remote --> Normalize --> Consumer
```

***

©️ 2026 [邱敬幃 Pardn Chiu](https://pardn.io)
