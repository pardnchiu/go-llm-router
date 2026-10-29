# OAuth Tokens

Where Copilot, Codex, and Grok tokens are stored, how to log in, and how refresh works.

## Keychain storage

Tokens for the three OAuth providers are stored by `go-pkg/filesystem/keychain` in the system keychain (macOS Keychain, Linux `secret-tool`).

| Provider | Keychain key | Login |
|---|---|---|
| Copilot | `COPILOT_OAUTH_TOKEN` | Device flow; `LoginWithCallback` yields a `*DeviceCode` to display |
| Codex | `CODEX_OAUTH_TOKEN` | PKCE; `LoginWithCallback` yields the authorization URL |
| Grok | `GROK_OAUTH_TOKEN` | PKCE, same shape |

All three fall back to the legacy keys `agenvoy.copilot.token`, `agenvoy.codex.token`, and `agenvoy.grok-oauth.token` on read, so existing installs need no new login.

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

`EnsureFresh` (Codex, Grok) refreshes the OAuth token within 60 seconds of expiry and writes the new token back to the keychain; `EnsureFreshSession` (Copilot) exchanges the stored GitHub token for a new short-lived session token within 60 seconds of the session's expiry, keeping it in memory only. Both run every time an agent builds its auth header, so a long-running process needs no refresh scheduler of its own.

## Related

- [OAuth Functions](/api-reference-oauth): signatures for `Load`, `LoginWithCallback`, `EnsureFresh`
