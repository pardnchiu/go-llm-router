# 訊息型別

請求訊息、工具與回應輸出型別的欄位。

## 型別

| 型別 | 欄位 |
|---|---|
| `Message` | `Role`、`Content`（`string` 或 `[]ContentPart`）、`ReasoningContent`、`ToolCalls`、`ToolCallID` |
| `ContentPart` | `Type`（`text` / `image_url`）、`Text`、`ImageURL` |
| `ImageURL` | `URL`（接受 data URI）、`Detail` |
| `Tool` / `ToolFunction` | `Type`、`Function{Name, Description, Parameters json.RawMessage}` |
| `ToolCall` | `ID`、`Type`、`Function`（`ToolCallFunction{Name, Arguments}`）、`ThoughtSignature` |
| `Output` | `Choices`、`Usage`、`ServiceTier`、`Error` |
| `OutputChoices` | `Message`、`Delta`、`FinishReason` |
| `Usage` | `Input`、`Output`、`CacheCreate`、`CacheRead`；自訂 `UnmarshalJSON` |

## 相關頁面

- [工具呼叫](/zh/core-concepts-tool-calls)
- [用量正規化](/zh/core-concepts-usage)
