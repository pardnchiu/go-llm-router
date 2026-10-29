# Audio Types

STT and TTS options, results, per-provider endpoints, and PCM helpers.

## Options and results

```go
type STTOptions struct{ Prompt, Language, MimeType string }
type STTResult struct{ Text string }
type TTSOptions struct{ Voice, Format string } // wav (default) / mp3 / opus / aac / flac / pcm
type TTSResult struct{ Audio []byte; MimeType string }
```

## Endpoints per provider

| Provider | STT endpoint | TTS endpoint | TTS result |
|---|---|---|---|
| `openai` | `/v1/audio/transcriptions` | `/v1/audio/speech` | requested `Format`, `wav` by default |
| `gemini` | `:generateContent` verbatim transcript | `:generateContent` `AUDIO` modality | WAV |
| `openrouter` | `/api/v1/audio/transcriptions` (multipart) | `/api/v1/audio/speech` | MP3 when `Format` is `mp3`, WAV otherwise |

## PCM helpers

| Function | Signature |
|---|---|
| `PCMRate` | `func(mime string) int`; falls back to `24000` |
| `WrapPCM16` | `func(pcm []byte, rate int) []byte` |

## Related

- [Speech Output](/core-concepts-speech): default voices and format handling
