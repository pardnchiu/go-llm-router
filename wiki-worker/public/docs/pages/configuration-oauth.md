# OAuth Tokens

Where Copilot, Codex, and Grok tokens are stored, how to log in, and how refresh works.

## Keychain storage

Tokens for the three OAuth providers are stored by `go-pkg/filesystem/keychain` in the system keychain (macOS Keychain, Linux `secret-tool`).

| Provider | Keychain key | Login |
|---|---|---|
| Copilot | `COPILOT_OAUTH_TOKEN` | Device flow; `LoginWithCallback` yields a `*DeviceCode` to display |
| Codex | `CODEX_OAUTH_TOKEN` | PKCE browser callback (`LoginWithCallback` yields the authorization URL) or device code (`LoginWithDeviceCode` yields a `*DeviceCode`); `Login` runs both and keeps the first to finish |
| Grok | `GROK_OAUTH_TOKEN` | PKCE browser callback; when the browser cannot reach the local callback, `Submit` takes the pasted code or callback URL |

All three fall back to the legacy keys `agenvoy.copilot.token`, `agenvoy.codex.token`, and `agenvoy.grok-oauth.token` on read, so existing installs need no new login.

## Logging in without a local browser

On a remote or headless machine the browser callback to `localhost` never arrives. Two paths cover that case:

| Provider | Path | Behavior |
|---|---|---|
| Codex | Device code | `onCode` receives `VerificationURI` (`https://auth.openai.com/codex/device`), `UserCode`, and `ExpiresAt`; the user enters the code on any device. An account without device code login gets `device code login is not enabled for this account; use the browser login` |
| Grok | Paste | After authorizing, the user copies the code or the whole redirected URL; `Submit` checks its `state` against the pending login and finishes it. A URL carrying `error` returns that error; a second submit for the same login, or a submit with no login waiting, is rejected |

Codex and Copilot device code polling share the same timing: the server's `interval` (falling back to `llmrouter.IntervalBackoff`, 5 seconds), stopping at `llmrouter.TimeoutWait` (10 minutes) or the server's expiry, whichever is earlier. A Copilot `slow_down` reply adds 5 seconds to the interval and polling continues.

## Using a stored token

```go
token, err := oauthCodex.Load()
if err != nil {
	return err
}
agent, err := router.New(router.Config{Name: "codex@gpt-5.1-codex", Token: token})
```

The packages under `core/oauth/copilot`, `core/oauth/codex`, and `core/oauth/grok` are named `oauthCopilot`, `oauthCodex`, and `oauthGrok`. `Load` returns `nil, nil` when nothing is stored, and `router.New` then fails with `Token is required`.

## Refresh

`EnsureFresh` (Codex, Grok) refreshes the OAuth token within `llmrouter.ExpirySafety` (60 seconds) of expiry and writes the new token back to the keychain; `EnsureFreshSession` (Copilot) exchanges the stored GitHub token for a new short-lived session token within 60 seconds of the session's expiry, keeping it in memory only. A Codex or Grok token response without `expires_in` is treated as valid for `llmrouter.ExpiryToken` (1 hour). Both run every time an agent builds its auth header, so a long-running process needs no refresh scheduler of its own.

## Related

- [OAuth Functions](/api-reference-oauth): signatures for `Load`, `Login`, `LoginWithCallback`, `LoginWithDeviceCode`, `Submit`, `EnsureFresh`
