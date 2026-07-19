package grokoauth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pardnchiu/go-llm-router/core"
	oauthGrok "github.com/pardnchiu/go-llm-router/core/oauth/grok"
)

const Prefix = "grok-oauth@"

func newHTTPClient() *http.Client {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		base = &http.Transport{}
	}
	transport := base.Clone()
	transport.ResponseHeaderTimeout = 15 * time.Second
	return &http.Client{
		Timeout:   10 * time.Minute,
		Transport: transport,
	}
}

type Agent struct {
	httpClient *http.Client
	model      string

	token *core.GrokToken
}

func New(config core.Config) (*Agent, error) {
	token, ok := config.Token.(*core.GrokToken)
	if !ok || token == nil {
		return nil, fmt.Errorf("grokoauth.New: Token is required")
	}

	return &Agent{
		httpClient: newHTTPClient(),
		model:      config.Model,
		token:      token,
	}, nil
}

func (a *Agent) Name() string {
	return Prefix + a.model
}

func (a *Agent) authHeader(ctx context.Context) (string, error) {
	token, err := oauthGrok.EnsureFresh(ctx, a.token)
	if err != nil {
		return "", fmt.Errorf("oauthGrokOauth.EnsureFresh: %w", err)
	}
	a.token = token
	return "Bearer " + token.AccessToken, nil
}
