# Message Types

Fields of the request message, tool, and response output types.

## Types

| Type | Fields |
|---|---|
| `Message` | `Role`, `Content` (`string` or `[]ContentPart`), `ReasoningContent`, `ToolCalls`, `ToolCallID` |
| `ContentPart` | `Type` (`text` / `image_url`), `Text`, `ImageURL` |
| `ImageURL` | `URL` (accepts a data URI), `Detail` |
| `Tool` / `ToolFunction` | `Type`, `Function{Name, Description, Parameters json.RawMessage}` |
| `ToolCall` | `ID`, `Type`, `Function` (`ToolCallFunction{Name, Arguments}`), `ThoughtSignature` |
| `Output` | `Choices`, `Usage`, `ServiceTier`, `Error` |
| `OutputChoices` | `Message`, `Delta`, `FinishReason` |
| `Usage` | `Input`, `Output`, `CacheCreate`, `CacheRead`; custom `UnmarshalJSON` |

## Related

- [Tool Calls](/core-concepts-tool-calls)
- [Usage Normalization](/core-concepts-usage)
