# 語音型別

STT 與 TTS 的選項、結果、各供應商端點與 PCM 輔助函式。

## 選項與結果

```go
type STTOptions struct{ Prompt, Language, MimeType string }
type STTResult struct{ Text string }
type TTSOptions struct{ Voice, Format string } // wav（預設）/ mp3 / opus / aac / flac / pcm
type TTSResult struct{ Audio []byte; MimeType string }
```

## 各供應商端點

| 供應商 | STT 端點 | TTS 端點 | TTS 輸出 |
|---|---|---|---|
| `openai` | `/v1/audio/transcriptions` | `/v1/audio/speech` | 依 `Format`，預設 `wav` |
| `gemini` | `:generateContent` 逐字轉寫 | `:generateContent` 的 `AUDIO` 模態 | WAV |
| `openrouter` | `/api/v1/audio/transcriptions`（multipart） | `/api/v1/audio/speech` | `Format` 為 `mp3` 時回 MP3，其餘回 WAV |

## PCM 輔助函式

| 函式 | 簽名 |
|---|---|
| `PCMRate` | `func(mime string) int`，預設 `24000` |
| `WrapPCM16` | `func(pcm []byte, rate int) []byte` |

## 相關頁面

- [語音輸出](/zh/core-concepts-speech)：預設 voice 與格式處理
