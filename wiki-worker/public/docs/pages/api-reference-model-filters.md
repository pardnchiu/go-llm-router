# Model Filter Symbols

Types and classifier functions behind `Models` filtering.

## Symbols

| Symbol | Signature |
|---|---|
| `ModelFilter` | `struct{ TextOnly, STTOnly, TTSOnly, ImageOnly bool }` |
| `ModelInfo` | `struct{ ID string; Thinking bool; Efforts, Endpoints []string }` |
| `IsTextModel` / `IsSTTModel` / `IsTTSModel` / `IsImageModel` | `func(id string) bool` |
| `MatchModelFilter` | `func(id string, filter ModelFilter) bool` |

## Listing response types

Wire shapes that `Models` decodes before filtering.

| Type | Used by | Fields |
|---|---|---|
| `Models` | `claude`, `openai`, `grok`, `grok-oauth`, `deepseek`, `nvidia`, `ollama-cloud`, `openrouter` | `Data[].ID` |
| `CopilotModels` | `copilot` | `Data[]`: `ID`, `ModelPickerEnabled`, `SupportedEndpoints`, `Capabilities.Type`, `Capabilities.Supports.ReasoningEffort`, `Policy.State` |
| `MistralModels` | `mistral` | `Data[]`: `ID`, `Capabilities.CompletionChat`, `Capabilities.Reasoning` |
| `GeminiModels` | `gemini` | `Models[]`: `Name`, `SupportedGenerationMethods`, `Thinking` |
| `CloudFlareModels` | `cloudflare` | `Result[]`: `Name`, `Task.Name` |

## Related

- [Model Filters](/core-concepts-model-filters): flag rules and per-provider support
