# Removed APIs

Exported symbols that no longer exist, the release that removed them, and what to use instead.

## v0.9.3

| Symbol | Package | Status | Last present | Replacement |
|---|---|---|---|---|
| `Usage` | `core/ollamaCloud` | Removed in v0.9.3 | v0.9.2 | no public replacement; Ollama Cloud reports spend instead of limits since its 2026-08-31 pricing change |

## v0.9.0

v0.9.0 moved provider names into shared constants in `core`; `Name()` output is unchanged.

| Symbol | Package | Status | Last present | Replacement |
|---|---|---|---|---|
| `Prefix` | `core/claude` | Removed in v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_CLAUDE + "@"` |
| `Prefix` | `core/cloudflare` | Removed in v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_CLOUDFLARE + "@"` |
| `Prefix` | `core/copilot` | Removed in v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_COPILOT + "@"` |
| `Prefix` | `core/deepseek` | Removed in v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_DEEPSEEK + "@"` |
| `Prefix` | `core/gemini` | Removed in v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_GEMINI + "@"` |
| `Prefix` | `core/grok` | Removed in v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_GROK + "@"` |
| `Prefix` | `core/grokOauth` | Removed in v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_GROK_OAUTH + "@"` |
| `Prefix` | `core/mistral` | Removed in v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_MISTRAL + "@"` |
| `Prefix` | `core/nvidia` | Removed in v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_NVIDIA + "@"` |
| `Prefix` | `core/ollamaCloud` | Removed in v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_OLLAMA_CLOUD + "@"` |
| `Prefix` | `core/openRouter` | Removed in v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_OPENROUTER + "@"` |
| `Prefix` | `core/openai` | Removed in v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_OPENAI + "@"` |
| `Prefix` | `core/openaiCodex` | Removed in v0.9.0 | v0.8.4 | `llmrouter.PROVIDER_CODEX + "@"` |

## v0.5.3

| Symbol | Package | Status | Last present | Replacement |
|---|---|---|---|---|
| `ImageOptions` | `core/openaiCodex` | Removed in v0.5.3 | v0.5.2 | `llmrouter.ImageOptions` through `ImageAgent.GenerateImage`, shared by every image provider |

## v0.3.0

v0.3.0 replaced string reasoning levels with the typed `llmrouter.Reasoning`; the string helpers were removed with it.

| Symbol | Package | Status | Last present | Replacement |
|---|---|---|---|---|
| `ReasoningDisabled` | `core` | Removed in v0.3.0 | v0.2.0 | compare with `ReasoningNone` |
| `ClampReasoningLevel` | `core` | Removed in v0.3.0 | v0.2.0 | `ClampReasoning` |
| `FloorReasoningLevel` | `core` | Removed in v0.3.0 | v0.2.0 | `ClampReasoning` |
| `MinReasoningLevel` | `core` | Removed in v0.3.0 | v0.2.0 | `ReasoningAgent.ReasoningLimits` (low bound) |
| `MaxReasoningLevel` | `core` | Removed in v0.3.0 | v0.2.0 | `ReasoningAgent.ReasoningLimits` (high bound) |
| `SupportReasoningEffort` | `core` | Removed in v0.3.0 | v0.2.0 | `ReasoningAgent` type assertion |
| `SupportsReasoningSwitch` | `core` | Removed in v0.3.0 | v0.2.0 | `ReasoningAgent` type assertion |
| `GetThinkingType` | `core` | Removed in v0.3.0 | v0.2.0 | no public replacement; each adapter maps `Reasoning` internally |
| `GetThinkingConfig` | `core` | Removed in v0.3.0 | v0.2.0 | no public replacement; each adapter maps `Reasoning` internally |
| `ThinkingBudget` | `core` | Removed in v0.3.0 | v0.2.0 | no public replacement; each adapter maps `Reasoning` internally |
| `RateLimitCooldown` | `core` | Removed in v0.3.0 | v0.2.0 | no public replacement |

## Related

- [Reasoning Levels](/core-concepts-reasoning)
- [Release Notes](/released/)
