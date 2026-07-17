# go-llm-router - 架構

> 返回 [README](./README.zh.md)

## 概覽

```mermaid
graph TB
    App[呼叫端應用] --> Router[core/router]
    Router --> Agent[core.Agent 介面]
    Agent --> Claude[claude]
    Agent --> OpenAI[openai]
    Agent --> Gemini[gemini]
    Agent --> Others[其他供應商]
    Agent --> Compat[compat]
    OAuth[core/oauth] --> Token[Token 物件]
    Token --> Router
    Agent --> Usage[正規化 Usage]
```

## 模組：core

定義統一合約與跨供應商共用邏輯。

```mermaid
graph TB
    subgraph core
        AgentI[Agent 介面]
        Types[Message / Tool / Output]
        UsageN[Usage.UnmarshalJSON]
        Reason[推理層級輔助函式]
        HTTP[NewHTTPClient]
    end
    AgentI --> Types
    Types --> UsageN
    Reason --> AgentI
    HTTP --> AgentI
```

## 模組：router

以 `provider@model` 字串解析並分派到對應工廠。

```mermaid
graph LR
    Config[router.Config] --> New[router.New]
    New --> Parse[解析 provider 與 model]
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

## 模組：供應商實作

每個供應商套件負責認證、請求塑形、回應正規化。

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
    Send --> HTTPAPI[供應商 HTTP API]
    HTTPAPI --> Output[core.Output]
```

## 模組：oauth

提供 device-code / refresh 流程，並可將 token 存入系統 keychain。

```mermaid
graph TB
    subgraph oauth
        CopilotO[oauth/copilot]
        CodexO[oauth/codex]
        GrokO[oauth/grok]
    end
    CopilotO --> KC[系統 Keychain]
    CodexO --> KC
    GrokO --> KC
    CopilotO --> CT[CopilotToken]
    CodexO --> XT[CodexToken]
    GrokO --> GT[GrokToken]
    CT --> Router[router.Config.Token]
    XT --> Router
    GT --> Router
```

## 資料流

```mermaid
sequenceDiagram
    participant App as 呼叫端
    participant Router as router.New
    participant Agent as core.Agent
    participant API as 供應商 API
    App->>Router: Config Name APIKey Token
    Router->>Agent: 建立供應商 Agent
    App->>Agent: Send messages tools reasoning
    Agent->>Agent: 夾住推理層級與請求塑形
    Agent->>API: HTTP 請求
    API-->>Agent: 供應商原始回應
    Agent-->>App: Output statusCode error
    Note over Agent,App: Usage 正規化為 Input Output CacheCreate CacheRead
```

## 狀態機：OAuth Token

```mermaid
stateDiagram-v2
    [*] --> Missing: 無 token
    Missing --> Pending: LoginWithCallback
    Pending --> Active: 裝置授權成功
    Active --> Active: Load 與使用
    Active --> Expired: Expired 為 true
    Expired --> Active: Refresh
    Active --> Missing: ClearToken
    Expired --> Missing: ClearToken
```

***

©️ 2026 [邱敬幃 Pardn Chiu](https://pardn.io)
