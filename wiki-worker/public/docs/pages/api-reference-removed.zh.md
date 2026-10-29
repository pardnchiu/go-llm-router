# 已移除 API

已不存在的匯出符號、移除它們的版本，以及替代做法。

## v0.9.3

| 符號 | 套件 | 狀態 | 最後存在版本 | 替代做法 |
|---|---|---|---|---|
| `Usage` | `core/ollamaCloud` | 移除於 v0.9.3 | v0.9.2 | 無公開替代；Ollama Cloud 自 2026-08-31 計價調整後改為回報花費而非額度上限 |

## v0.9.0

v0.9.0 把供應商名稱收斂為 `core` 的共用常數；`Name()` 輸出不變。

| 符號 | 套件 | 狀態 | 最後存在版本 | 替代做法 |
|---|---|---|---|---|
| `Prefix` | `core/claude` | 移除於 v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_CLAUDE + "@"` |
| `Prefix` | `core/cloudflare` | 移除於 v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_CLOUDFLARE + "@"` |
| `Prefix` | `core/copilot` | 移除於 v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_COPILOT + "@"` |
| `Prefix` | `core/deepseek` | 移除於 v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_DEEPSEEK + "@"` |
| `Prefix` | `core/gemini` | 移除於 v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_GEMINI + "@"` |
| `Prefix` | `core/grok` | 移除於 v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_GROK + "@"` |
| `Prefix` | `core/grokOauth` | 移除於 v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_GROK_OAUTH + "@"` |
| `Prefix` | `core/mistral` | 移除於 v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_MISTRAL + "@"` |
| `Prefix` | `core/nvidia` | 移除於 v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_NVIDIA + "@"` |
| `Prefix` | `core/ollamaCloud` | 移除於 v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_OLLAMA_CLOUD + "@"` |
| `Prefix` | `core/openRouter` | 移除於 v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_OPENROUTER + "@"` |
| `Prefix` | `core/openai` | 移除於 v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_OPENAI + "@"` |
| `Prefix` | `core/openaiCodex` | 移除於 v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_CODEX + "@"` |

## v0.5.3

| 符號 | 套件 | 狀態 | 最後存在版本 | 替代做法 |
|---|---|---|---|---|
| `ImageOptions` | `core/openaiCodex` | 移除於 v0.5.3 | v0.5.2 | 透過 `ImageAgent.GenerateImage` 使用所有圖片供應商共用的 `llmrouter.ImageOptions` |

## v0.3.0

v0.3.0 以型別化的 `llmrouter.Reasoning` 取代字串形式的推理等級，字串輔助函式一併移除。

| 符號 | 套件 | 狀態 | 最後存在版本 | 替代做法 |
|---|---|---|---|---|
| `ReasoningDisabled` | `core` | 移除於 v0.3.0 | v0.2.0 | 與 `ReasoningNone` 比較 |
| `ClampReasoningLevel` | `core` | 移除於 v0.3.0 | v0.2.0 | `ClampReasoning` |
| `FloorReasoningLevel` | `core` | 移除於 v0.3.0 | v0.2.0 | `ClampReasoning` |
| `MinReasoningLevel` | `core` | 移除於 v0.3.0 | v0.2.0 | `ReasoningAgent.ReasoningLimits`（下限） |
| `MaxReasoningLevel` | `core` | 移除於 v0.3.0 | v0.2.0 | `ReasoningAgent.ReasoningLimits`（上限） |
| `SupportReasoningEffort` | `core` | 移除於 v0.3.0 | v0.2.0 | `ReasoningAgent` 型別斷言 |
| `SupportsReasoningSwitch` | `core` | 移除於 v0.3.0 | v0.2.0 | `ReasoningAgent` 型別斷言 |
| `GetThinkingType` | `core` | 移除於 v0.3.0 | v0.2.0 | 無公開替代；各轉接層內部對映 `Reasoning` |
| `GetThinkingConfig` | `core` | 移除於 v0.3.0 | v0.2.0 | 無公開替代；各轉接層內部對映 `Reasoning` |
| `ThinkingBudget` | `core` | 移除於 v0.3.0 | v0.2.0 | 無公開替代；各轉接層內部對映 `Reasoning` |
| `RateLimitCooldown` | `core` | 移除於 v0.3.0 | v0.2.0 | 無公開替代 |

## 相關頁面

- [推理等級](/zh/core-concepts-reasoning)
- [版本紀錄](/released/)
