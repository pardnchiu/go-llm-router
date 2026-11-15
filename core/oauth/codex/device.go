package oauthCodex

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

func LoginWithDeviceCode(ctx context.Context, onCode func(*DeviceCode)) (*llmrouter.CodexToken, error) {
	var issued struct {
		DeviceAuthID string `json:"device_auth_id"`
		UserCode     string `json:"user_code"`
		Interval     string `json:"interval"`
		ExpiresAt    string `json:"expires_at"`
	}
	status, err := postDeviceJSON(ctx, deviceUserCodeURL, map[string]string{"client_id": clientID}, &issued)
	if err != nil {
		return nil, fmt.Errorf("request user code: %w", err)
	}
	switch {
	case status == http.StatusNotFound:
		return nil, fmt.Errorf("device code login is not enabled for this account; use the browser login")
	case status != http.StatusOK:
		return nil, fmt.Errorf("request user code: status %d", status)
	case issued.DeviceAuthID == "" || issued.UserCode == "":
		return nil, fmt.Errorf("request user code: empty device_auth_id or user_code")
	}

	interval := llmrouter.IntervalBackoff
	if n, err := strconv.Atoi(strings.TrimSpace(issued.Interval)); err == nil && n > 0 {
		interval = time.Duration(n) * time.Second
	}
	deadline := time.Now().Add(llmrouter.TimeoutWait)
	if t, err := time.Parse(time.RFC3339Nano, issued.ExpiresAt); err == nil && t.Before(deadline) {
		deadline = t
	}

	if onCode != nil {
		onCode(&DeviceCode{
			VerificationURI: deviceVerifyURL,
			UserCode:        issued.UserCode,
			ExpiresAt:       deadline,
		})
	}

	body := map[string]string{"device_auth_id": issued.DeviceAuthID, "user_code": issued.UserCode}
	for {
		var granted struct {
			AuthorizationCode string `json:"authorization_code"`
			CodeVerifier      string `json:"code_verifier"`
		}
		status, err := postDeviceJSON(ctx, deviceTokenURL, body, &granted)
		if err != nil {
			return nil, fmt.Errorf("poll device token: %w", err)
		}
		switch {
		case status == http.StatusOK:
			if granted.AuthorizationCode == "" || granted.CodeVerifier == "" {
				return nil, fmt.Errorf("poll device token: empty authorization_code or code_verifier")
			}
			return exchangeCode(ctx, granted.AuthorizationCode, granted.CodeVerifier, deviceRedirectURI)
		case status != http.StatusForbidden && status != http.StatusNotFound:
			return nil, fmt.Errorf("poll device token: status %d", status)
		}

		if time.Now().After(deadline) {
			return nil, fmt.Errorf("device code expired: %w", context.DeadlineExceeded)
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(min(interval, time.Until(deadline)+time.Second)):
		}
	}
}

func postDeviceJSON(ctx context.Context, endpoint string, payload any, out any) (int, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("json.Marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return 0, fmt.Errorf("http.NewRequestWithContext: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := llmrouter.NewStreamHTTPClient().Do(req)
	if err != nil {
		return 0, fmt.Errorf("httpClient.Do: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
	if err != nil {
		return resp.StatusCode, fmt.Errorf("io.ReadAll: %w", err)
	}
	if resp.StatusCode == http.StatusOK {
		if err := json.Unmarshal(data, out); err != nil {
			return resp.StatusCode, fmt.Errorf("json.Unmarshal: %w", err)
		}
	}
	return resp.StatusCode, nil
}

func Login(ctx context.Context, onURL func(string), onCode func(*DeviceCode)) (*llmrouter.CodexToken, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type result struct {
		path  string
		token *llmrouter.CodexToken
		err   error
	}
	results := make(chan result, 2)
	go func() {
		token, err := LoginWithCallback(ctx, onURL)
		results <- result{"browser login", token, err}
	}()
	go func() {
		token, err := LoginWithDeviceCode(ctx, onCode)
		results <- result{"device code login", token, err}
	}()

	var errs []error
	for range 2 {
		r := <-results
		if r.err == nil {
			return r.token, nil
		}
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		errs = append(errs, fmt.Errorf("%s: %w", r.path, r.err))
	}
	return nil, errors.Join(errs...)
}
