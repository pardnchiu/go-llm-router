# OAuth Functions

The function sets of `core/oauth/copilot`, `core/oauth/codex`, and `core/oauth/grok`.

## Functions

| Function | Copilot (`oauthCopilot`) | Codex / Grok |
|---|---|---|
| `Load` | `func() (*llmrouter.CopilotToken, error)` | `func() (*llmrouter.CodexToken, error)` / `func() (*llmrouter.GrokToken, error)` |
| `HasToken` | `func() bool` | `func() bool` |
| `ClearToken` | `func() error` | `func() error` |
| `LoginWithCallback` | `func(ctx context.Context, onCode func(*DeviceCode)) (*llmrouter.CopilotToken, error)` | `func(ctx context.Context, onURL func(string)) (*llmrouter.CodexToken, error)` / `(*llmrouter.GrokToken, error)` |
| `Login` | — | Codex only: `func(ctx context.Context, onURL func(string), onCode func(*DeviceCode)) (*llmrouter.CodexToken, error)`; runs the browser callback and device code logins at once and returns the first token, or both errors joined |
| `LoginWithDeviceCode` | — | Codex only: `func(ctx context.Context, onCode func(*DeviceCode)) (*llmrouter.CodexToken, error)` |
| `Submit` | — | Grok only: `func(input string) error`; completes a pending `LoginWithCallback` with a pasted authorization code or full callback URL |
| `EnsureFresh` | — | `func(ctx context.Context, token *llmrouter.CodexToken) (*llmrouter.CodexToken, error)` / `GrokToken` equivalent |
| `EnsureFreshSession` | `func(ctx context.Context, token *llmrouter.CopilotToken, refresh *llmrouter.CopilotRefreshToken) (*llmrouter.CopilotRefreshToken, error)` | — |

## Copilot device-flow types

| Type | Fields |
|---|---|
| `DeviceCode` | `DeviceCode`, `UserCode`, `VerificationURI`, `ExpiresIn`, `Interval`; passed to `onCode` so the caller can show `UserCode` and `VerificationURI` |
| `GopilotAccessToken` | `AccessToken`, `TokenType`, `Scope`, `Error`; the access-token poll response of the device flow |

## Codex device-code type

| Type | Fields |
|---|---|
| `DeviceCode` (`oauthCodex`) | `VerificationURI`, `UserCode`, `ExpiresAt`; passed to `onCode` so the caller can show where to enter `UserCode` and until when |

## Token types

The token types `CopilotToken`, `CopilotRefreshToken`, `CodexToken`, and `GrokToken` live in `core`; `CodexToken` and `GrokToken` expose `Expired() bool`.

## Related

- [OAuth Tokens](/configuration-oauth): keychain keys and login flows
