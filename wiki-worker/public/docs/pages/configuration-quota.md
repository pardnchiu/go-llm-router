# Balance and Quota

Which provider packages expose a `Usage` function and which `UsageRemaining` fields each one fills.

## Signature

Every package below exports the same function:

```go
func Usage(ctx context.Context, config llmrouter.Config) (llmrouter.UsageRemaining, error)
```

```go
type UsageRemaining struct {
	FiveHour *float64 `json:"five_hour,omitempty"` // percentage left in the 5-hour window
	Week     *float64 `json:"week,omitempty"`      // percentage left in the weekly window
	Balance  *float64 `json:"balance,omitempty"`   // account balance or credits left
	Total    *float64 `json:"total,omitempty"`     // percentage left of the overall quota
}
```

Each provider sets only the fields it reports; the rest stay `nil`, so check a field before dereferencing it.

## Fields per package

| Package | Credentials | Fields set |
|---|---|---|
| `core/openRouter` | `APIKey` | `Balance` = `total_credits - total_usage` |
| `core/deepseek` | `APIKey` | `Balance` = first `balance_infos[].total_balance` |
| `core/copilot` | `Token` | `Total` = `percent_remaining` of `chat`, or of `premium_interactions` when the chat entitlement is 0 |
| `core/openaiCodex` | `APIKey`, optional `AccountID` | `FiveHour` / `Week` = `100 - used_percent` of the primary / secondary rate-limit window |
| `core/grokOauth` | `APIKey` | `Total` = remaining credit percentage |
| `core/claudeCode` | none (uses the logged-in `claude` CLI) | `FiveHour` / `Week` = `100 -` the `Current session` / `Current week (all models)` percentage from `claude -p /usage` |

The Go package names for `core/openRouter`, `core/openaiCodex`, and `core/grokOauth` are `openrouter`, `openaicodex`, and `grokoauth`. A missing credential returns an error before any request is sent. `claudeCode.Usage` returns an error when the CLI is missing or its output has no subscription limits. `Usage` in `core/ollamaCloud` was removed in v0.9.3; see [Removed APIs](/api-reference-removed).

## Example

```go
remaining, err := openaicodex.Usage(ctx, llmrouter.Config{APIKey: token.AccessToken, AccountID: token.AccountID})
if err != nil {
	return err
}
if remaining.FiveHour != nil && remaining.Week != nil {
	fmt.Printf("5h: %.0f%% left, week: %.0f%% left\n", *remaining.FiveHour, *remaining.Week)
}
```

## Related

- [Usage Normalization](/core-concepts-usage): per-request token counters
- [Interfaces and Router](/api-reference#provider-packages): per-package exports
