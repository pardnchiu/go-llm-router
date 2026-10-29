# 圖片型別

圖片選項、結果、各供應商端點與圖片輔助函式。

## 選項與結果

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

## 各供應商端點

| 供應商 | 端點 | 結果 |
|---|---|---|
| `openai` | `/v1/images/generations`，帶參考圖時 `/v1/images/edits` | PNG（`output_format`），含 `Revised` |
| `codex` | ChatGPT Codex Responses 的 `image_generation` 工具 | 依結果事件的 `output_format`，缺省為 PNG；含 `Revised` |
| `gemini` | `:generateContent` | 依回應的 `inlineData.mimeType` |
| `grok` / `grok-oauth` | `/v1/images/generations`，帶參考圖時 `/v1/images/edits` | 依回應的 `mime_type`，缺省為 JPEG |
| `openrouter` | `/api/v1/images` | 依回應的 `media_type`，缺省為 PNG |

## 輔助函式

| 函式 | 簽名 |
|---|---|
| `ImagePixelSize` | `func(opts ImageOptions) string`，`Size` 設定短邊 |
| `DataURI` | `func(mime, b64 string) string` |

## 相關頁面

- [圖片生成](/zh/core-concepts-images)
