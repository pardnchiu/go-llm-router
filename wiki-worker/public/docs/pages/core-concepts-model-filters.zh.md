# 模型篩選

如何列出供應商的模型，並篩出對話、語音合成、轉寫或圖片模型。

## 列出模型

```go
ids, err := openai.Models(ctx, llmrouter.Config{APIKey: key}, llmrouter.ModelFilter{TextOnly: true})
```

## 旗標規則

`ModelFilter` 的四個旗標同時成立時取交集。以標記判斷的供應商對模型 ID 套用下列規則：

| 旗標 | 判準 |
|---|---|
| `TextOnly` | 排除圖片／語音／影片／embedding 等非對話模型 |
| `STTOnly` | 含 `transcribe`、`whisper-` 標記 |
| `TTSOnly` | 含 `-tts`、`tts-` 標記 |
| `ImageOnly` | 含 `-image`、`image-`、`imagen-`、`imagine-`、`nano-banana` 標記，且**排除** `-video`、`video-`、`veo-`、`sora-` |

`STTOnly`、`TTSOnly`、`ImageOnly` 另外排除含 `-live` 的 ID，這類模型不走一般 HTTP。

`ImageOnly` 特地排除 video 的理由：`grok-imagine-video` 同時命中 `imagine-` 標記，不排除就會出現在圖片模型清單裡，讓使用者選到一個 `GenerateImage` 無法使用的模型。

## 供應商支援

各供應商支援的旗標不同，不支援的旗標會被忽略而非報錯：

| 供應商 | `TextOnly` | `STTOnly` | `TTSOnly` | `ImageOnly` | 來源 |
|---|---|---|---|---|---|
| `openai`、`gemini`、`grok`、`grok-oauth` | ✓ | ✓ | ✓ | ✓ | 上述模型 ID 標記 |
| `openrouter` | ✓ | ✓ | ✓ | ✓ | 供應商目錄，每次只取一個（`ImageOnly` > `TTSOnly` > `STTOnly`）：`/api/v1/images/models`、`/api/v1/models?output_modalities=speech` / `transcription`；`TextOnly` 再套上述標記 |
| `cloudflare` | ✓ | ✓ | ✓ | — | Workers AI task 名稱（`STTOnly` > `TTSOnly`；預設 `Text Generation`） |
| `claude`、`copilot`、`deepseek`、`mistral`、`nvidia`、`ollama-cloud`、`codex` | ✓ | — | — | — | 上述模型 ID 標記 |
| `claude-code` | — | — | — | — | `claudeCode.Models(ctx, config)` 不接受篩選條件，回傳 CLI `initialize` 回應中解析後的模型 ID |

## 相關頁面

- [模型篩選符號](/zh/api-reference-model-filters)：`ModelFilter`、`ModelInfo` 與分類函式
