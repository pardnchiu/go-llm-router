# go-llm-router - Architecture

> Back to [README](../README.md)

## Overview

```mermaid
graph TB
    App[Caller App] --> Router[core/router]
    Router --> Agent[core.Agent Interface]
    Agent --> Claude[claude]
    Agent --> OpenAI[openai]
    Agent --> Gemini[gemini]
    Agent --> Others[Other Providers]
    Agent --> Compat[compat]
    OAuth[core/oauth] --> Token[Token Objects]
    Token --> Router
    Agent --> Usage[Normalized Usage]
```

## Module: core

Defines the shared contract and cross-provider helpers.

```mermaid
graph TB
    subgraph core
        AgentI[Agent Interface]
        Types[Message / Tool / Output]
        UsageN[Usage.UnmarshalJSON]
        Reason[Reasoning Helpers]
        HTTP[NewHTTPClient]
    end
    AgentI --> Types
    Types --> UsageN
    Reason --> AgentI
    HTTP --> AgentI
```

## Module: router

Parses `provider@model` and dispatches to the matching factory.

```mermaid
graph LR
    Config[router.Config] --> New[router.New]
    New --> Parse[Parse provider and model]
    Parse --> Map[newFn map]
    Map --> Claude[claude.New]
    Map --> OpenAI[openai.New]
    Map --> Gemini[gemini.New]
    Map --> OAuthP[copilot / codex / grok-oauth]
    Map --> Compat[compat.New]
    Claude --> Agent[core.Agent]
    OpenAI --> Agent
    Gemini --> Agent
    OAuthP --> Agent
    Compat --> Agent
```

## Module: Provider Implementations

Each provider package owns auth, request shaping, and response normalization.

```mermaid
graph TB
    subgraph Providers
        Claude[claude]
        OpenAI[openai]
        Codex[openaiCodex]
        Gemini[gemini]
        Grok[grok / grokOauth]
        DeepSeek[deepseek]
        NVIDIA[nvidia]
        OpenRouter[openRouter]
        Cloudflare[cloudflare]
        Copilot[copilot]
        Compat[compat]
    end
    Providers --> Send[Send]
    Send --> HTTPAPI[Provider HTTP API]
    HTTPAPI --> Output[core.Output]
```

## Module: oauth

Provides device-code and refresh flows, optionally backed by the system keychain.

```mermaid
graph TB
    subgraph oauth
        CopilotO[oauth/copilot]
        CodexO[oauth/codex]
        GrokO[oauth/grok]
    end
    CopilotO --> KC[System Keychain]
    CodexO --> KC
    GrokO --> KC
    CopilotO --> CT[CopilotToken]
    CodexO --> XT[CodexToken]
    GrokO --> GT[GrokToken]
    CT --> Router[router.Config.Token]
    XT --> Router
    GT --> Router
```

## Data Flow

```mermaid
sequenceDiagram
    participant App as Caller
    participant Router as router.New
    participant Agent as core.Agent
    participant API as Provider API
    App->>Router: Config Name APIKey Token
    Router->>Agent: Build provider Agent
    App->>Agent: Send messages tools reasoning
    Agent->>Agent: Clamp reasoning and shape request
    Agent->>API: HTTP request
    API-->>Agent: Provider raw response
    Agent-->>App: Output statusCode error
    Note over Agent,App: Usage normalized to Input Output CacheCreate CacheRead
```

## State Machine: OAuth Token

```mermaid
stateDiagram-v2
    [*] --> Missing: no token
    Missing --> Pending: LoginWithCallback
    Pending --> Active: device authorization succeeds
    Active --> Active: Load and use
    Active --> Expired: Expired is true
    Expired --> Active: Refresh
    Active --> Missing: ClearToken
    Expired --> Missing: ClearToken
```

***

©️ 2026 [邱敬幃 Pardn Chiu](https://pardn.io)
