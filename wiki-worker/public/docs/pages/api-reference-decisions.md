# Decision Types

Decision request and result types, the shared request builder and response decoder, and the endpoint each provider calls.

## Interface

```go
type DecisionAgent interface {
	Decide(ctx context.Context, state any, questions map[string]DecisionQuestion) (*DecisionResult, error)
}
```

Implemented by `*typesafe.Agent` and `*openrouter.Agent`.

## Question

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

## Result

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

| Field | Set for | Wire field |
|---|---|---|
| `Noul` | `noul` | `answers.<name>.noul` |
| `Choice` | `choice` | `answers.<name>.choice` |
| `Score` | `score` | `answers.<name>.score` |
| `Confidence` | `choice`, `score` | `answers.<name>.confidence` |
| `Probabilities` | `choice` (by option), `score` (by level index) | `answers.<name>.probabilities` |
| `Legend` | `score` | `answers.<name>.legend` |
| `Usage.Input` / `Usage.Output` | all | `usage.input_tokens` / `usage.output_tokens` |
| `Usage.CacheRead` | `openrouter` only | `usage.input_tokens_details.cached_tokens` |

## Shared helpers

Both adapters build the request and decode the response with the same functions in `core`.

| Symbol | Signature | Description |
|---|---|---|
| `DecisionBody` | `func(model string, state any, questions map[string]DecisionQuestion) (map[string]any, error)` | Validates and returns `{model, state, questions}`; nil `Instructions` / `Criteria` are omitted. Returns `[Invalid] decision...` on an empty model, nil `state`, empty `questions`, or an unknown `Type` |
| `DecisionResponse` | `struct` | Wire shape of both upstream responses: `model`, `answers`, `usage` |
| `(DecisionResponse).Result` | `func(label string) (*DecisionResult, error)` | Converts to `DecisionResult`; returns `[Upstream] <label>: response has no answers` when `answers` is empty |
| `HTTPErrorLabel` | `func(code int, err error) string` | Error label for an HTTP failure, see below |

## HTTPErrorLabel

The adapters prefix every upstream error with this label, so a caller can branch on the category.

| Condition | Label |
|---|---|
| `err` wraps `context.Canceled` | `Canceled` |
| `err` wraps `context.DeadlineExceeded` | `Timeout` |
| `code == 0` (no response) | `Unavailable` |
| 400, 415, 422 | `Invalid` |
| 401, 403 | `Denied` |
| 402, 413 | `Limit` |
| 404 | `Not found` |
| 409 | `Conflict` |
| 408, 504, 524 | `Timeout` |
| 429, 503, 529 | `Unavailable` |
| anything else, including a 200 whose body fails to decode | `Upstream` |

## Endpoints per provider

| Provider | Package | Endpoint | Headers |
|---|---|---|---|
| `typesafe` | `core/typesafe` | `POST https://api.typesafe.ai/v1/systemone` | `Authorization: Bearer <APIKey>` |
| `openrouter` | `core/openRouter` | `POST https://openrouter.ai/api/alpha/decisions` | the same headers as OpenRouter chat |

## core/typesafe

| Symbol | Signature | Description |
|---|---|---|
| `Agent` | `struct` | Decision-only agent; it has no `Send` and is not an `llmrouter.Agent` |
| `New` | `func(config llmrouter.Config) (*Agent, error)` | Uses `Model` and `APIKey`; an empty `APIKey` returns `[Denied] typesafe.New: APIKey is required` |
| `(*Agent).Name` | `func() string` | `typesafe@<model>` |
| `(*Agent).Decide` | `DecisionAgent` | Calls `/v1/systemone` |
| `Models` | `func(ctx context.Context, config llmrouter.Config) ([]string, error)` | Names from `GET /v1/models`, 10-second timeout; takes no `ModelFilter` |

## Related

- [Decisions](/core-concepts-decisions): question types, provider rules, and examples
- [Interfaces and Router](/api-reference): `router.NewDecision` and `PROVIDER_TYPESAFE`
