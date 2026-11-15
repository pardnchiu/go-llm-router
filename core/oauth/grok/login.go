package oauthGrok

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

var pending struct {
	mu    sync.Mutex
	state string
	codes chan string
}

func Submit(input string) error {
	code, state, err := parsePasted(input)
	if err != nil {
		return err
	}

	pending.mu.Lock()
	defer pending.mu.Unlock()
	if pending.codes == nil {
		return fmt.Errorf("no grok login is waiting for a code")
	}
	if state != "" && state != pending.state {
		return fmt.Errorf("state mismatch: the pasted URL belongs to another login")
	}
	select {
	case pending.codes <- code:
		return nil
	default:
		return fmt.Errorf("a code was already submitted for this login")
	}
}

func parsePasted(input string) (code, state string, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", fmt.Errorf("empty input")
	}
	if u, perr := url.Parse(input); perr == nil && u.Scheme != "" {
		q := u.Query()
		if e := q.Get("error"); e != "" {
			return "", "", fmt.Errorf("%s: %s", e, q.Get("error_description"))
		}
		if q.Get("code") == "" {
			return "", "", fmt.Errorf("URL has no code parameter")
		}
		return q.Get("code"), q.Get("state"), nil
	}
	if strings.ContainsAny(input, " \t\r\n") {
		return "", "", fmt.Errorf("not an authorization code or callback URL")
	}
	return input, "", nil
}

func LoginWithCallback(ctx context.Context, onURL func(string)) (*llmrouter.GrokToken, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}
	verifier := base64.RawURLEncoding.EncodeToString(b)
	sum := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(sum[:])

	b = make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate state: %w", err)
	}
	state := base64.RawURLEncoding.EncodeToString(b)

	b = make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonce := base64.RawURLEncoding.EncodeToString(b)

	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	srv, err := startCallbackServer(state, codeCh, errCh)
	if err != nil {
		return nil, fmt.Errorf("startCallbackServer: %w", err)
	}
	defer srv.Shutdown(context.Background())

	pasted := make(chan string, 1)
	pending.mu.Lock()
	pending.state, pending.codes = state, pasted
	pending.mu.Unlock()
	defer func() {
		pending.mu.Lock()
		if pending.codes == pasted {
			pending.state, pending.codes = "", nil
		}
		pending.mu.Unlock()
	}()

	authLink := buildAuthURL(challenge, state, nonce)

	if onURL != nil {
		onURL(authLink)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, fmt.Errorf("callback: %w", err)
	case code := <-codeCh:
		return exchangeCode(ctx, code, verifier, challenge)
	case code := <-pasted:
		return exchangeCode(ctx, code, verifier, challenge)
	}
}

func exchangeCode(ctx context.Context, code, verifier, challenge string) (*llmrouter.GrokToken, error) {
	form := url.Values{
		"grant_type":            {"authorization_code"},
		"code":                  {code},
		"redirect_uri":          {redirectURI},
		"client_id":             {clientID},
		"code_verifier":         {verifier},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL,
		strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := llmrouter.NewStreamHTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("httpClient.Do: %w", err)
	}
	defer resp.Body.Close()

	var raw oauthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("json.Decode: %w", err)
	}
	if raw.Error != nil {
		return nil, fmt.Errorf("token error %v: %v", raw.Error, raw.ErrorDesc)
	}

	expiry := time.Now().Add(time.Duration(raw.ExpiresIn) * time.Second)
	if raw.ExpiresIn == 0 {
		expiry = time.Now().Add(llmrouter.ExpiryToken)
	}

	token := &llmrouter.GrokToken{
		AccessToken:  raw.AccessToken,
		RefreshToken: raw.RefreshToken,
		ExpiresAt:    expiry,
	}

	if err := saveToken(token); err != nil {
		return nil, fmt.Errorf("saveToken: %w", err)
	}
	return token, nil
}

func buildAuthURL(challenge, state, nonce string) string {
	v := url.Values{
		"response_type":         {"code"},
		"client_id":             {clientID},
		"redirect_uri":          {redirectURI},
		"scope":                 {scopes},
		"state":                 {state},
		"nonce":                 {nonce},
		"code_challenge":        {challenge},
		"code_challenge_method": {"S256"},
		"plan":                  {"generic"},
	}
	return authURL + "?" + v.Encode()
}

var xaiCORSAllowed = map[string]bool{
	"https://accounts.x.ai": true,
	"https://auth.x.ai":     true,
}

func startCallbackServer(expectedState string, codeCh chan<- string, errCh chan<- error) (*http.Server, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:56121")
	if err != nil {
		return nil, fmt.Errorf("net.Listen: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if xaiCORSAllowed[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Private-Network", "true")
			w.Header().Set("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		q := r.URL.Query()

		if errParam := q.Get("error"); errParam != "" {
			desc := q.Get("error_description")
			fmt.Fprintf(w, "Authorization failed: %s: %s", errParam, desc)
			errCh <- fmt.Errorf("%s: %s", errParam, desc)
			return
		}

		if q.Get("state") != expectedState {
			fmt.Fprint(w, "Authorization failed: state mismatch")
			errCh <- fmt.Errorf("state mismatch")
			return
		}

		code := q.Get("code")
		if code == "" {
			fmt.Fprint(w, "Authorization failed: missing code")
			errCh <- fmt.Errorf("missing code")
			return
		}

		fmt.Fprint(w, "Authorization successful — you can close this tab.")
		codeCh <- code
	})

	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  llmrouter.TimeoutHTTPShort,
		WriteTimeout: llmrouter.TimeoutHTTPShort,
	}
	go srv.Serve(listener)
	return srv, nil
}
