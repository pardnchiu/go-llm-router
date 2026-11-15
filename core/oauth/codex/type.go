package oauthCodex

import "time"

const (
	tokenKey          = "CODEX_OAUTH_TOKEN"
	clientID          = "app_EMoamEEZ73f0CkXaXp7hrann"
	endpoint          = "https://auth.openai.com"
	authURL           = endpoint + "/oauth/authorize"
	tokenURL          = endpoint + "/oauth/token"
	redirectURI       = "http://localhost:1455/auth/callback"
	scopes            = "openid email profile offline_access"
	deviceUserCodeURL = endpoint + "/api/accounts/deviceauth/usercode"
	deviceTokenURL    = endpoint + "/api/accounts/deviceauth/token"
	deviceVerifyURL   = endpoint + "/codex/device"
	deviceRedirectURI = endpoint + "/deviceauth/callback"
)

type oauthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	IDToken      string `json:"id_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	// * Error / ErrorDesc are `any` because OpenAI's token endpoint switched
	// * `error` from string to object (e.g. {"code":"...","message":"..."}).
	Error     any `json:"error,omitempty"`
	ErrorDesc any `json:"error_description,omitempty"`
}

type DeviceCode struct {
	VerificationURI string
	UserCode        string
	ExpiresAt       time.Time
}
