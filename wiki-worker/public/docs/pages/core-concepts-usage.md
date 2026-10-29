# Usage Normalization

How vendor-specific token counters fold into one `Usage` shape.

## The Usage shape

Vendors name token fields differently: the OpenAI family uses `prompt_tokens` / `completion_tokens`, Anthropic uses `input_tokens` / `output_tokens`, and cache fields diverge further. `Usage.UnmarshalJSON` folds them into one shape:

```go
type Usage struct {
	Input       int
	Output      int
	CacheCreate int
	CacheRead   int
}
```

## Cache accounting

`prompt_tokens_details.cached_tokens` is subtracted from `Input` and folded into `CacheRead`, so a cache hit is never counted twice. Responses-API adapters (`input_tokens_details.cached_tokens`) and Gemini (`cachedContentTokenCount`) apply the same subtraction.

## Related

- [Balance and Quota](/configuration-quota): account-level remaining credit
- [Message Types](/api-reference-messages): where `Usage` sits in `Output`
