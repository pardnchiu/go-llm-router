# OAuth 權杖

Copilot／Codex／Grok 權杖存在哪裡、如何登入、如何刷新。

## Keychain 儲存

三家 OAuth 供應商的權杖由 `go-pkg/filesystem/keychain` 保存在系統 keychain（macOS Keychain、Linux `secret-tool`）。

| 供應商 | Keychain 鍵 | 登入方式 |
|---|---|---|
| Copilot | `COPILOT_OAUTH_TOKEN` | device flow，`LoginWithCallback` 回傳 `*DeviceCode` 供顯示驗證碼 |
| Codex | `CODEX_OAUTH_TOKEN` | PKCE，`LoginWithCallback` 回傳授權 URL |
| Grok | `GROK_OAUTH_TOKEN` | PKCE，同上 |

三者讀取時皆相容舊鍵 `agenvoy.copilot.token`、`agenvoy.codex.token`、`agenvoy.grok-oauth.token`，既有安裝不需重新登入。

## 使用已保存的權杖

```go
token, err := oauthCodex.Load()
if err != nil {
	return err
}
agent, err := router.New(router.Config{Name: "codex@gpt-5.1-codex", Token: token})
```

`core/oauth/copilot`、`core/oauth/codex`、`core/oauth/grok` 的套件名分別為 `oauthCopilot`、`oauthCodex`、`oauthGrok`。keychain 中沒有權杖時 `Load` 回傳 `nil, nil`，之後 `router.New` 會回 `Token is required`。

## 刷新

`EnsureFresh`（Codex、Grok）在 OAuth 權杖距過期 60 秒內換新，並把新權杖寫回 keychain；`EnsureFreshSession`（Copilot）在 session 距過期 60 秒內以保存的 GitHub 權杖換取新的短期 session token，只保留在記憶體。兩者在 Agent 每次組 auth header 時都會呼叫，因此長時間執行的程序不需要自行排程刷新。

## 相關頁面

- [OAuth 函式](/zh/api-reference-oauth)：`Load`、`LoginWithCallback`、`EnsureFresh` 簽名
