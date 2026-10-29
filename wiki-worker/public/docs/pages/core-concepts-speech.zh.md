# 語音輸出

各 TTS 供應商的預設 voice，以及能回傳哪些音訊格式。

## Voice 與格式

語音模型的選法與對話模型相同；輸出取決於各上游能編碼的格式：

| 供應商 | 預設 voice | `TTSOptions.Format` |
|---|---|---|
| `openai` | `alloy` | 原樣轉送；留空為 `wav` |
| `gemini` | `Kore` | 忽略；一律由 PCM 封裝為 WAV |
| `openrouter` | OpenRouter 目錄中該模型 `supported_voices` 的第一個 | `mp3` 直接回傳；其他值一律請求 PCM 並回傳 WAV |

## OpenRouter voice 驗證

OpenRouter 的 voice 依模型而異、沒有共用預設值，因此不支援的 voice 會在本地直接回錯並列出該模型的合法清單，取代上游語意不明的 `Provider returned 400`。

## 相關頁面

- [語音型別](/zh/api-reference-audio)：STT 與 TTS 選項、端點、PCM 輔助函式
- [模型篩選](/zh/core-concepts-model-filters)：以 `TTSOnly`／`STTOnly` 列出語音模型
