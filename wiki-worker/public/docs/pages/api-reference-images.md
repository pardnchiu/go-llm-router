# Image Types

Image options, results, per-provider endpoints, and image helpers.

## Options and result

```go
type ImageOptions struct {
	AspectRatio string // "1:1" "16:9" "4:3"
	Size        string // "1k" "2k" "4k"
	Quality     string // "low" "medium" "high"
	RefImageB64 string
	RefMime     string
}

type ImageResult struct {
	B64      string
	MimeType string
	Revised  string
}
```

## Endpoints per provider

| Provider | Endpoint | Result |
|---|---|---|
| `openai` | `/v1/images/generations`, `/v1/images/edits` with a reference | PNG (`output_format`), `Revised` set |
| `codex` | `image_generation` tool on ChatGPT Codex Responses | `output_format` of the result event, PNG when absent; `Revised` set |
| `gemini` | `:generateContent` | `inlineData.mimeType` from the response |
| `grok` / `grok-oauth` | `/v1/images/generations`, `/v1/images/edits` with a reference | `mime_type` from the response, JPEG when absent |
| `openrouter` | `/api/v1/images` | `media_type` from the response, PNG when absent |

## Helpers

| Function | Signature |
|---|---|
| `ImagePixelSize` | `func(opts ImageOptions) string`; `Size` applies to the short edge |
| `DataURI` | `func(mime, b64 string) string` |

## Related

- [Image Generation](/core-concepts-images)
