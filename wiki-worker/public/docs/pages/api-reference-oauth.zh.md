# OAuth 函式

`core/oauth/copilot`、`core/oauth/codex`、`core/oauth/grok` 的函式組。

## 函式

| 函式 | Copilot（`oauthCopilot`） | Codex / Grok |
|---|---|---|
| `Load` | `func() (*llmrouter.CopilotToken, error)` | `func() (*llmrouter.CodexToken, error)` / `func() (*llmrouter.GrokToken, error)` |
| `HasToken` | `func() bool` | `func() bool` |
| `ClearToken` | `func() error` | `func() error` |
| `LoginWithCallback` | `func(ctx context.Context, onCode func(*DeviceCode)) (*llmrouter.CopilotToken, error)` | `func(ctx context.Context, onURL func(string)) (*llmrouter.CodexToken, error)` / `(*llmrouter.GrokToken, error)` |
| `Login` | — | 僅 Codex：`func(ctx context.Context, onURL func(string), onCode func(*DeviceCode)) (*llmrouter.CodexToken, error)`；同時跑瀏覽器 callback 與 device code 登入，回傳先取得的權杖，兩者皆失敗時合併兩個錯誤 |
| `LoginWithDeviceCode` | — | 僅 Codex：`func(ctx context.Context, onCode func(*DeviceCode)) (*llmrouter.CodexToken, error)` |
| `Submit` | — | 僅 Grok：`func(input string) error`；以貼上的授權碼或完整 callback URL 完成進行中的 `LoginWithCallback` |
| `EnsureFresh` | — | `func(ctx context.Context, token *llmrouter.CodexToken) (*llmrouter.CodexToken, error)` / `GrokToken` 版本 |
| `EnsureFreshSession` | `func(ctx context.Context, token *llmrouter.CopilotToken, refresh *llmrouter.CopilotRefreshToken) (*llmrouter.CopilotRefreshToken, error)` | — |

## Copilot device flow 型別

| 型別 | 欄位 |
|---|---|
| `DeviceCode` | `DeviceCode`、`UserCode`、`VerificationURI`、`ExpiresIn`、`Interval`；傳給 `onCode`，供呼叫端顯示 `UserCode` 與 `VerificationURI` |
| `GopilotAccessToken` | `AccessToken`、`TokenType`、`Scope`、`Error`；device flow 輪詢 access token 的回應 |

## Codex device code 型別

| 型別 | 欄位 |
|---|---|
| `DeviceCode`（`oauthCodex`） | `VerificationURI`、`UserCode`、`ExpiresAt`；傳給 `onCode`，供呼叫端顯示 `UserCode` 的輸入位置與有效期限 |

## 權杖型別

權杖型別 `CopilotToken`、`CopilotRefreshToken`、`CodexToken`、`GrokToken` 位於 `core`；`CodexToken` 與 `GrokToken` 有 `Expired() bool`。

## 相關頁面

- [OAuth 權杖](/zh/configuration-oauth)：keychain 鍵與登入流程
