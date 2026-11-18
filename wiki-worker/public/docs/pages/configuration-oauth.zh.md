# OAuth 權杖

Copilot／Codex／Grok 權杖存在哪裡、如何登入、如何刷新。

## Keychain 儲存

三家 OAuth 供應商的權杖由 `go-pkg/filesystem/keychain` 保存在系統 keychain（macOS Keychain、Linux `secret-tool`）。

| 供應商 | Keychain 鍵 | 登入方式 |
|---|---|---|
| Copilot | `COPILOT_OAUTH_TOKEN` | device flow，`LoginWithCallback` 回傳 `*DeviceCode` 供顯示驗證碼 |
| Codex | `CODEX_OAUTH_TOKEN` | PKCE 瀏覽器 callback（`LoginWithCallback` 回傳授權 URL）或 device code（`LoginWithDeviceCode` 回傳 `*DeviceCode`）；`Login` 同時跑兩者，採用先完成者 |
| Grok | `GROK_OAUTH_TOKEN` | PKCE 瀏覽器 callback；瀏覽器連不到本機 callback 時，以 `Submit` 傳入貼上的授權碼或 callback URL |

三者讀取時皆相容舊鍵 `agenvoy.copilot.token`、`agenvoy.codex.token`、`agenvoy.grok-oauth.token`，既有安裝不需重新登入。

## 無本機瀏覽器時登入

在遠端或無介面的機器上，瀏覽器導回 `localhost` 的 callback 永遠不會抵達。兩條路徑處理這種情況：

| 供應商 | 路徑 | 行為 |
|---|---|---|
| Codex | Device code | `onCode` 收到 `VerificationURI`（`https://auth.openai.com/codex/device`）、`UserCode`、`ExpiresAt`，使用者可在任一裝置輸入代碼。帳號未開放 device code 登入時回 `device code login is not enabled for this account; use the browser login` |
| Grok | 貼上 | 授權後使用者複製授權碼或整個導回後的 URL，`Submit` 比對其 `state` 與進行中的登入後完成登入。URL 帶 `error` 時回傳該錯誤；同一次登入重複送出，或沒有等待中的登入時送出，皆會被拒絕 |

Codex 與 Copilot 的 device code 輪詢採相同時序：使用伺服器回傳的 `interval`（缺值時退回 `llmrouter.IntervalBackoff`，5 秒），於 `llmrouter.TimeoutWait`（10 分鐘）或伺服器到期時間兩者較早者停止。Copilot 收到 `slow_down` 時間隔加 5 秒並繼續輪詢。

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

`EnsureFresh`（Codex、Grok）在 OAuth 權杖距過期 `llmrouter.ExpirySafety`（60 秒）內換新，並把新權杖寫回 keychain；`EnsureFreshSession`（Copilot）在 session 距過期 60 秒內以保存的 GitHub 權杖換取新的短期 session token，只保留在記憶體。Codex 或 Grok 的權杖回應未帶 `expires_in` 時，視為有效 `llmrouter.ExpiryToken`（1 小時）。兩者在 Agent 每次組 auth header 時都會呼叫，因此長時間執行的程序不需要自行排程刷新。

## 相關頁面

- [OAuth 函式](/zh/api-reference-oauth)：`Load`、`Login`、`LoginWithCallback`、`LoginWithDeviceCode`、`Submit`、`EnsureFresh` 簽名
