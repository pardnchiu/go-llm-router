# Decisions

How to ask a model typed questions about a piece of text and read back one probability-weighted answer per question.

## Picking a decision agent

A decision call sends a `state` (the content to judge) and a set of named questions, and gets back structured answers instead of generated text. Two providers serve it, both built with `router.NewDecision`:

| Name | Upstream | Model example |
|---|---|---|
| `typesafe@<model>` | `POST https://api.typesafe.ai/v1/systemone` | `typesafe@jev-latest` |
| `openrouter@<model>` | `POST https://openrouter.ai/api/alpha/decisions` | `openrouter@typesafe/jev-1.13` |

```go
agent, err := router.NewDecision(router.Config{Name: "typesafe@jev-latest", APIKey: key})
```

`router.NewDecision` accepts only these two prefixes; any other prefix returns `[Invalid] router.NewDecision: ...`. The `openrouter` chat agent from `router.New` also implements `llmrouter.DecisionAgent`, so an existing OpenRouter agent can be asserted instead of built twice:

```go
decider, ok := agent.(llmrouter.DecisionAgent)
```

## Asking questions

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

`state` is any JSON-encodable value: a string, a map of related fields, or a slice. Question names are chosen by the caller and come back as the keys of `result.Answers`.

## Question types

| `Type` | `Criteria` | Answer fields |
|---|---|---|
| `DecisionNoul` | optional; `map[string]string{"true": ..., "false": ...}` describing what counts as yes and no | `Noul` in 0–1: near 1 favors true, near 0 favors false, near 0.5 is uncertain |
| `DecisionChoice` | required; map of option name → description (an empty description is judged by the name alone) | `Choice` (highest-probability option), `Confidence`, `Probabilities` per option |
| `DecisionScore` | required; ordered slice of levels, index 0 first | `Score` (probability-weighted level, may be fractional), `Confidence`, `Legend` (level index → criteria), `Probabilities` per level |

`Instructions` and `Criteria` are `any` and are sent as written, so a string, a map, or a slice all work. `Confidence` is a pointer because the noul answer carries none.

## Writing questions that work on both providers

Both providers accept the same request shape, but validate it differently:

| Rule | `typesafe` | `openrouter` |
|---|---|---|
| `Instructions` on every question | optional | required |
| noul `Criteria` | `true` and `false` each optional | when present, both keys required |
| score levels | 1–10 | at least 1 |
| choice options | 1–255 | at least 1 |
| `result.Usage.CacheRead` | always 0 | cached input tokens |
| `result.Model` | concrete version, e.g. `jev-1.13.0` | concrete version, e.g. `typesafe/jev-1.13-20260917` |

Giving every question `Instructions`, sending noul `Criteria` with both keys or not at all, and keeping scores to 10 levels lets one `questions` map run on either provider unchanged.

## Errors

`Decide` checks the request before sending it: an empty model, a nil `state`, an empty `questions` map, or a `Type` other than `noul` / `choice` / `score` returns `[Invalid] decision...` without a network call. Upstream failures are labeled by `HTTPErrorLabel` and keep the upstream body:

```text
[Denied] typesafe systemone: HTTP 401: {"detail":{"error_type":"authentication_error",...}}
[Denied] openrouter decisions: HTTP 401: {"error":{"message":"Missing Authentication header","code":401}}
[Canceled] typesafe systemone: failed to send: ... context canceled
```

A 200 response without any answer returns `[Upstream] <label>: response has no answers`.

## Listing decision models

```go
names, err := typesafe.Models(ctx, llmrouter.Config{APIKey: key})
ids, err := openrouter.Models(ctx, llmrouter.Config{APIKey: key}, llmrouter.ModelFilter{DecisionOnly: true})
```

`typesafe.Models` returns the model names and aliases from `GET /v1/models`; `openrouter.Models` with `DecisionOnly` reads `/api/v1/models?output_modalities=decisions`.

## Related

- [Decision Types](/api-reference-decisions): every type, field, and wire mapping
- [Model Filters](/core-concepts-model-filters): `DecisionOnly` among the other flags
- [Provider Credentials](/configuration): the API key for `typesafe@`
