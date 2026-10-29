# 模型篩選符號

`Models` 篩選背後的型別與分類函式。

## 符號

| 符號 | 簽名 |
|---|---|
| `ModelFilter` | `struct{ TextOnly, STTOnly, TTSOnly, ImageOnly bool }` |
| `ModelInfo` | `struct{ ID string; Thinking bool; Efforts, Endpoints []string }` |
| `IsTextModel` / `IsSTTModel` / `IsTTSModel` / `IsImageModel` | `func(id string) bool` |
| `MatchModelFilter` | `func(id string, filter ModelFilter) bool` |

## 模型清單回應型別

`Models` 在篩選前解碼的傳輸形狀。

| 型別 | 使用者 | 欄位 |
|---|---|---|
| `Models` | `claude`、`openai`、`grok`、`grok-oauth`、`deepseek`、`nvidia`、`ollama-cloud`、`openrouter` | `Data[].ID` |
| `CopilotModels` | `copilot` | `Data[]`：`ID`、`ModelPickerEnabled`、`SupportedEndpoints`、`Capabilities.Type`、`Capabilities.Supports.ReasoningEffort`、`Policy.State` |
| `MistralModels` | `mistral` | `Data[]`：`ID`、`Capabilities.CompletionChat`、`Capabilities.Reasoning` |
| `GeminiModels` | `gemini` | `Models[]`：`Name`、`SupportedGenerationMethods`、`Thinking` |
| `CloudFlareModels` | `cloudflare` | `Result[]`：`Name`、`Task.Name` |

## 相關頁面

- [模型篩選](/zh/core-concepts-model-filters)：旗標規則與各供應商支援
