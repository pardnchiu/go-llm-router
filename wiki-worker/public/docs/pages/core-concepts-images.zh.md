# 圖片生成

如何選圖片模型，並以共用的 `ImageOptions` 呼叫 `GenerateImage`。

## 選擇圖片模型

圖片與語音模型就是 agent 模型，選法與對話模型完全一致 —— `openai@gpt-image-2`、`gemini@gemini-3.1-flash-image`、`grok@grok-imagine-image-2.0`、`openrouter@google/gemini-2.5-flash-image`。唯一例外是 `codex`：ChatGPT 後端以對話模型本身生成圖片，沒有可選的圖片模型。

## 生成

```go
imageAgent, ok := agent.(llmrouter.ImageAgent)
if !ok {
	return fmt.Errorf("%s 無法生成圖片", agent.Name())
}
result, err := imageAgent.GenerateImage(ctx, prompt, llmrouter.ImageOptions{AspectRatio: "16:9", Size: "1k"})
```

## 選項對映

`ImageOptions` 是共用形狀，各家自行對映：不支援的選項會被丟棄而不是報錯。`Size` 依供應商對映：`openai` 換算成像素，`Size` 設定的是**短邊**，`16:9` + `1k` 得到 `1824x1024`；`gemini` 與 `openrouter` 轉大寫送出（`1K`）；`grok` 與 `grok-oauth` 送 `1k` 或 `2k`，`4k` 會壓成 `2k`。`gemini` 忽略 `Quality`，`codex` 忽略 `AspectRatio`、`Size` 與 `Quality`。

## 相關頁面

- [圖片型別](/zh/api-reference-images)：各供應商端點與輸出格式
- [模型篩選](/zh/core-concepts-model-filters)：以 `ImageOnly` 列出圖片模型
