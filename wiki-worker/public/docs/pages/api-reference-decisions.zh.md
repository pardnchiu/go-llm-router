# 決策型別

決策的請求與結果型別、共用的請求組裝與回應解析函式，以及各供應商呼叫的端點。

## 介面

```go
type DecisionAgent interface {
	Decide(ctx context.Context, state any, questions map[string]DecisionQuestion) (*DecisionResult, error)
}
```

由 `*typesafe.Agent` 與 `*openrouter.Agent` 實作。

## 題目

```go
type DecisionType string

const (
	DecisionNoul   DecisionType = "noul"
	DecisionChoice DecisionType = "choice"
	DecisionScore  DecisionType = "score"
)

type DecisionQuestion struct {
	Type         DecisionType
	Instructions any
	Criteria     any
}
```

## 結果

```go
type DecisionAnswer struct {
	Type          DecisionType
	Noul          float64
	Choice        string
	Score         float64
	Confidence    *float64
	Probabilities map[string]float64
	Legend        map[string]any
}

type DecisionResult struct {
	Model   string
	Answers map[string]DecisionAnswer
	Usage   Usage
}
```

| 欄位 | 適用題型 | Wire 欄位 |
|---|---|---|
| `Noul` | `noul` | `answers.<name>.noul` |
| `Choice` | `choice` | `answers.<name>.choice` |
| `Score` | `score` | `answers.<name>.score` |
| `Confidence` | `choice`、`score` | `answers.<name>.confidence` |
| `Probabilities` | `choice`（依選項）、`score`（依等級索引） | `answers.<name>.probabilities` |
| `Legend` | `score` | `answers.<name>.legend` |
| `Usage.Input`／`Usage.Output` | 全部 | `usage.input_tokens`／`usage.output_tokens` |
| `Usage.CacheRead` | 僅 `openrouter` | `usage.input_tokens_details.cached_tokens` |

## 共用函式

兩個 adapter 以 `core` 內同一組函式組裝請求與解析回應。

| 符號 | 簽章 | 說明 |
|---|---|---|
| `DecisionBody` | `func(model string, state any, questions map[string]DecisionQuestion) (map[string]any, error)` | 驗證後回傳 `{model, state, questions}`；值為 nil 的 `Instructions`／`Criteria` 不送出。model 為空、`state` 為 nil、`questions` 為空或 `Type` 不合法時回傳 `[Invalid] decision...` |
| `DecisionResponse` | `struct` | 兩個上游回應共用的 wire 格式：`model`、`answers`、`usage` |
| `(DecisionResponse).Result` | `func(label string) (*DecisionResult, error)` | 轉成 `DecisionResult`；`answers` 為空時回傳 `[Upstream] <label>: response has no answers` |
| `HTTPErrorLabel` | `func(code int, err error) string` | HTTP 失敗對應的錯誤標示，見下節 |

## HTTPErrorLabel

adapter 在每個上游錯誤前加上這個標示，呼叫端可依類別分流處理。

| 條件 | 標示 |
|---|---|
| `err` 包含 `context.Canceled` | `Canceled` |
| `err` 包含 `context.DeadlineExceeded` | `Timeout` |
| `code == 0`（沒有回應） | `Unavailable` |
| 400、415、422 | `Invalid` |
| 401、403 | `Denied` |
| 402、413 | `Limit` |
| 404 | `Not found` |
| 409 | `Conflict` |
| 408、504、524 | `Timeout` |
| 429、503、529 | `Unavailable` |
| 其他，包含 200 但回應無法解析 | `Upstream` |

## 各供應商端點

| 供應商 | 套件 | 端點 | Header |
|---|---|---|---|
| `typesafe` | `core/typesafe` | `POST https://api.typesafe.ai/v1/systemone` | `Authorization: Bearer <APIKey>` |
| `openrouter` | `core/openRouter` | `POST https://openrouter.ai/api/alpha/decisions` | 與 OpenRouter 對話相同的 header |

## core/typesafe

| 符號 | 簽章 | 說明 |
|---|---|---|
| `Agent` | `struct` | 只做決策的 agent；沒有 `Send`，不是 `llmrouter.Agent` |
| `New` | `func(config llmrouter.Config) (*Agent, error)` | 使用 `Model` 與 `APIKey`；`APIKey` 為空時回傳 `[Denied] typesafe.New: APIKey is required` |
| `(*Agent).Name` | `func() string` | `typesafe@<model>` |
| `(*Agent).Decide` | `DecisionAgent` | 呼叫 `/v1/systemone` |
| `Models` | `func(ctx context.Context, config llmrouter.Config) ([]string, error)` | `GET /v1/models` 的模型名稱，逾時 10 秒；不接受 `ModelFilter` |

## 相關頁面

- [決策問答](/zh/core-concepts-decisions)：題型、供應商規則與範例
- [介面與工廠](/zh/api-reference)：`router.NewDecision` 與 `PROVIDER_TYPESAFE`
