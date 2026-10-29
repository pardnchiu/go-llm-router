# Speech Output

How each TTS provider picks a default voice and which audio formats it can return.

## Voices and formats

Speech models are picked like chat models; output depends on what each upstream can encode:

| Provider | Default voice | `TTSOptions.Format` |
|---|---|---|
| `openai` | `alloy` | forwarded as-is; empty means `wav` |
| `gemini` | `Kore` | ignored; always WAV built from PCM |
| `openrouter` | first `supported_voices` entry of the model in the OpenRouter catalog | `mp3` passes through; every other value requests PCM and returns WAV |

## OpenRouter voice validation

OpenRouter voices are model-specific and share no common default, so an unsupported voice fails locally with the model's valid list instead of the upstream's opaque `Provider returned 400`.

## Related

- [Audio Types](/api-reference-audio): STT and TTS options, endpoints, PCM helpers
- [Model Filters](/core-concepts-model-filters): listing speech models with `TTSOnly` / `STTOnly`
