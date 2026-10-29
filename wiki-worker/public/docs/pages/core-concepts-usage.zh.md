# 用量正規化

各家不同的 token 計數欄位如何折疊成單一 `Usage` 形狀。

## Usage 形狀

各家的 token 欄位命名不同：OpenAI 系列用 `prompt_tokens` / `completion_tokens`，Anthropic 用 `input_tokens` / `output_tokens`，快取欄位更是各走各的。`Usage.UnmarshalJSON` 把它們折疊成同一個形狀：

```go
type Usage struct {
	Input       int
	Output      int
	CacheCreate int
	CacheRead   int
}
```

## 快取計算

`prompt_tokens_details.cached_tokens` 會自 `Input` 扣除後計入 `CacheRead`，避免快取命中的部分被重複計價。Responses API 轉接層（`input_tokens_details.cached_tokens`）與 Gemini（`cachedContentTokenCount`）採用相同的扣除方式。

## 相關頁面

- [餘額與配額](/zh/configuration-quota)：帳戶層級的剩餘額度
- [訊息型別](/zh/api-reference-messages)：`Usage` 在 `Output` 中的位置
