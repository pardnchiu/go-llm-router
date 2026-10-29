# Model Filters

How to list a provider's models and narrow them to text, speech, transcription, or image models.

## Listing models

```go
ids, err := openai.Models(ctx, llmrouter.Config{APIKey: key}, llmrouter.ModelFilter{TextOnly: true})
```

## Flag rules

The four `ModelFilter` flags intersect. Marker-based providers apply these rules to model IDs:

| Flag | Rule |
|---|---|
| `TextOnly` | drops image, audio, video, embedding, and other non-chat models |
| `STTOnly` | keeps `transcribe`, `whisper-` markers |
| `TTSOnly` | keeps `-tts`, `tts-` markers |
| `ImageOnly` | keeps `-image`, `image-`, `imagen-`, `imagine-`, `nano-banana` markers and **excludes** `-video`, `video-`, `veo-`, `sora-` |

`STTOnly`, `TTSOnly`, and `ImageOnly` also drop any ID containing `-live`, since those models are not served over plain HTTP.

`ImageOnly` excludes video deliberately: `grok-imagine-video` also matches the `imagine-` marker, and without the exclusion it would show up in an image-model picker as a model `GenerateImage` cannot use.

## Provider support

Flag support differs per provider; an unsupported flag is ignored rather than rejected:

| Provider | `TextOnly` | `STTOnly` | `TTSOnly` | `ImageOnly` | Source |
|---|---|---|---|---|---|
| `openai`, `gemini`, `grok`, `grok-oauth` | ✓ | ✓ | ✓ | ✓ | model-ID markers above |
| `openrouter` | ✓ | ✓ | ✓ | ✓ | vendor catalogs, one per call (`ImageOnly` > `TTSOnly` > `STTOnly`): `/api/v1/images/models`, `/api/v1/models?output_modalities=speech` / `transcription`; `TextOnly` adds the markers above |
| `cloudflare` | ✓ | ✓ | ✓ | — | Workers AI task name (`STTOnly` > `TTSOnly`; default `Text Generation`) |
| `claude`, `copilot`, `deepseek`, `mistral`, `nvidia`, `ollama-cloud`, `codex` | ✓ | — | — | — | model-ID markers above |
| `claude-code` | — | — | — | — | `claudeCode.Models(ctx, config)` takes no filter; it returns the resolved model IDs from the CLI's `initialize` response |

## Related

- [Model Filter Symbols](/api-reference-model-filters): `ModelFilter`, `ModelInfo`, classifier functions
