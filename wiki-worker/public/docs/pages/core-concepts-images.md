# Image Generation

How to pick an image model and call `GenerateImage` with the shared `ImageOptions`.

## Picking an image model

Image and audio models are the agent model, picked exactly like chat models — `openai@gpt-image-2`, `gemini@gemini-3.1-flash-image`, `grok@grok-imagine-image-2.0`, `openrouter@google/gemini-2.5-flash-image`. The one exception is `codex`: the ChatGPT backend generates from the chat model itself, so there is no separate image model to pick.

## Generating

```go
imageAgent, ok := agent.(llmrouter.ImageAgent)
if !ok {
	return fmt.Errorf("%s cannot generate images", agent.Name())
}
result, err := imageAgent.GenerateImage(ctx, prompt, llmrouter.ImageOptions{AspectRatio: "16:9", Size: "1k"})
```

## Options mapping

`ImageOptions` is one shared shape each adapter maps onto its own wire parameters; unsupported options are dropped rather than rejected. `Size` maps per provider: `openai` converts it to pixels with `Size` as the **short** edge, so `16:9` + `1k` becomes `1824x1024`; `gemini` and `openrouter` send it uppercased (`1K`); `grok` and `grok-oauth` send `1k` or `2k`, with `4k` clamped to `2k`. `gemini` drops `Quality`, and `codex` ignores `AspectRatio`, `Size`, and `Quality`.

## Related

- [Image Types](/api-reference-images): endpoints and result formats per provider
- [Model Filters](/core-concepts-model-filters): listing image models with `ImageOnly`
