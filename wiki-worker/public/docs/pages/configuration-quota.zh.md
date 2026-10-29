# 餘額與配額

哪些供應商套件提供 `Usage` 函式，以及各自填入 `UsageRemaining` 的哪些欄位。

## 簽名

下列套件都匯出同一個函式：

```go
func Usage(ctx context.Context, config llmrouter.Config) (llmrouter.UsageRemaining, error)
```

```go
type UsageRemaining struct {
	FiveHour *float64 `json:"five_hour,omitempty"` // 5 小時視窗剩餘百分比
	Week     *float64 `json:"week,omitempty"`      // 週視窗剩餘百分比
	Balance  *float64 `json:"balance,omitempty"`   // 帳戶餘額或剩餘點數
	Total    *float64 `json:"total,omitempty"`     // 整體配額剩餘百分比
}
```

各供應商只填入自家回報的欄位，其餘維持 `nil`，取值前須先判斷。

## 各套件填入欄位

| 套件 | 憑證 | 填入欄位 |
|---|---|---|
| `core/openRouter` | `APIKey` | `Balance` = `total_credits - total_usage` |
| `core/deepseek` | `APIKey` | `Balance` = 第一筆 `balance_infos[].total_balance` |
| `core/copilot` | `Token` | `Total` = `chat` 的 `percent_remaining`；chat entitlement 為 0 時改取 `premium_interactions` |
| `core/openaiCodex` | `APIKey`，`AccountID` 選填 | `FiveHour` / `Week` = 主／次 rate-limit 視窗的 `100 - used_percent` |
| `core/grokOauth` | `APIKey` | `Total` = 剩餘點數百分比 |
| `core/claudeCode` | 無（使用已登入的 `claude` CLI） | `FiveHour` / `Week` = `100 -` `claude -p /usage` 中 `Current session`／`Current week (all models)` 的百分比 |

`core/openRouter`、`core/openaiCodex`、`core/grokOauth` 的 Go 套件名分別為 `openrouter`、`openaicodex`、`grokoauth`。缺少憑證時在送出請求前直接回錯。`claudeCode.Usage` 在 CLI 不存在或輸出中沒有訂閱額度時回錯。`core/ollamaCloud` 的 `Usage` 已於 v0.9.3 移除，見 [已移除 API](/zh/api-reference-removed)。

## 範例

```go
remaining, err := openaicodex.Usage(ctx, llmrouter.Config{APIKey: token.AccessToken, AccountID: token.AccountID})
if err != nil {
	return err
}
if remaining.FiveHour != nil && remaining.Week != nil {
	fmt.Printf("5h: %.0f%% left, week: %.0f%% left\n", *remaining.FiveHour, *remaining.Week)
}
```

## 相關頁面

- [用量正規化](/zh/core-concepts-usage)：單次請求的 token 計數
- [介面與工廠](/zh/api-reference#供應商套件)：各套件匯出
