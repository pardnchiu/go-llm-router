# 決策問答

對一段文字提出型別化問題，每題取回一個以機率加權的答案。

## 選擇決策 agent

決策呼叫送出 `state`（要判斷的內容）與一組具名題目，取回結構化答案而不是生成文字。目前由兩個供應商提供，都以 `router.NewDecision` 建立：

| 名稱 | 上游 | 模型範例 |
|---|---|---|
| `typesafe@<model>` | `POST https://api.typesafe.ai/v1/systemone` | `typesafe@jev-latest` |
| `openrouter@<model>` | `POST https://openrouter.ai/api/alpha/decisions` | `openrouter@typesafe/jev-1.13` |

```go
agent, err := router.NewDecision(router.Config{Name: "typesafe@jev-latest", APIKey: key})
```

`router.NewDecision` 只接受這兩個前綴，其他前綴回傳 `[Invalid] router.NewDecision: ...`。`router.New` 建出的 `openrouter` 對話 agent 同樣實作 `llmrouter.DecisionAgent`，已有 OpenRouter agent 時直接斷言即可，不必再建一次：

```go
decider, ok := agent.(llmrouter.DecisionAgent)
```

## 提問

```go
result, err := agent.Decide(ctx,
	"Our API started returning 500 errors 20 minutes ago and we cannot process orders.",
	map[string]llmrouter.DecisionQuestion{
		"department": {
			Type:         llmrouter.DecisionChoice,
			Instructions: "Which team should handle this",
			Criteria: map[string]string{
				"billing":   "Payment or subscription issues",
				"technical": "Bugs or integration problems",
				"sales":     "Pricing or account questions",
			},
		},
		"frustration": {
			Type:         llmrouter.DecisionScore,
			Instructions: "How frustrated the customer appears",
			Criteria:     []string{"Calm, just stating facts", "Frustrated but civil", "Very angry"},
		},
		"is_urgent": {
			Type:         llmrouter.DecisionNoul,
			Instructions: "The message conveys urgency or time-sensitivity",
		},
	})
if err != nil {
	return err
}
fmt.Println(result.Answers["department"].Choice, result.Answers["is_urgent"].Noul)
```

`state` 可以是任何能編碼成 JSON 的值：字串、相關欄位組成的 map，或 slice。題目名稱由呼叫端自訂，回傳時即為 `result.Answers` 的 key。

## 題型

| `Type` | `Criteria` | 答案欄位 |
|---|---|---|
| `DecisionNoul` | 選填；`map[string]string{"true": ..., "false": ...}`，說明什麼算是、什麼算否 | `Noul` 介於 0–1：接近 1 偏向 true、接近 0 偏向 false、接近 0.5 代表不確定 |
| `DecisionChoice` | 必填；選項名稱 → 說明的 map（說明留空時只依名稱判斷） | `Choice`（機率最高的選項）、`Confidence`、每個選項的 `Probabilities` |
| `DecisionScore` | 必填；依序排列的等級 slice，索引 0 為第一級 | `Score`（機率加權後的等級，可為小數）、`Confidence`、`Legend`（等級索引 → criteria）、每個等級的 `Probabilities` |

`Instructions` 與 `Criteria` 型別為 `any`，原樣送出，字串、map、slice 都可以。`Confidence` 是指標，因為 noul 答案不帶信心值。

## 兩個供應商通用的寫法

兩邊接受同一種請求格式，但驗證規則不同：

| 規則 | `typesafe` | `openrouter` |
|---|---|---|
| 每題的 `Instructions` | 選填 | 必填 |
| noul 的 `Criteria` | `true`、`false` 各自選填 | 有給時兩個 key 都必填 |
| score 等級數 | 1–10 | 至少 1 |
| choice 選項數 | 1–255 | 至少 1 |
| `result.Usage.CacheRead` | 一律為 0 | 快取命中的輸入 token |
| `result.Model` | 具體版本，例如 `jev-1.13.0` | 具體版本，例如 `typesafe/jev-1.13-20260917` |

每題都給 `Instructions`、noul 的 `Criteria` 兩個 key 都給或都不給、score 不超過 10 級，同一份 `questions` 就能在兩個供應商上直接使用。

## 錯誤

`Decide` 送出前先檢查請求：model 為空、`state` 為 nil、`questions` 為空，或 `Type` 不是 `noul`／`choice`／`score`，都直接回傳 `[Invalid] decision...`，不發出網路請求。上游錯誤由 `HTTPErrorLabel` 標示類別，並保留上游回應內容：

```text
[Denied] typesafe systemone: HTTP 401: {"detail":{"error_type":"authentication_error",...}}
[Denied] openrouter decisions: HTTP 401: {"error":{"message":"Missing Authentication header","code":401}}
[Canceled] typesafe systemone: failed to send: ... context canceled
```

回應為 200 但沒有任何答案時，回傳 `[Upstream] <label>: response has no answers`。

## 列出決策模型

```go
names, err := typesafe.Models(ctx, llmrouter.Config{APIKey: key})
ids, err := openrouter.Models(ctx, llmrouter.Config{APIKey: key}, llmrouter.ModelFilter{DecisionOnly: true})
```

`typesafe.Models` 回傳 `GET /v1/models` 的模型名稱與別名；`openrouter.Models` 搭配 `DecisionOnly` 讀取 `/api/v1/models?output_modalities=decisions`。

## 相關頁面

- [決策型別](/zh/api-reference-decisions)：所有型別、欄位與 wire 對應
- [模型篩選](/zh/core-concepts-model-filters)：`DecisionOnly` 與其他旗標
- [供應商憑證](/zh/configuration)：`typesafe@` 的 API key
