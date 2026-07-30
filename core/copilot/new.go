package copilot

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/pardnchiu/go-llm-router/core"
	oauthCopilot "github.com/pardnchiu/go-llm-router/core/oauth/copilot"
)

type Agent struct {
	httpClient *http.Client
	model      string
	Token      *core.CopilotToken
	Refresh    *core.CopilotRefreshToken
	efforts    []string
	endpoints  []string

	endpointMu    sync.Mutex
	endpointDone  bool
	endpointCache []string
}

const (
	Prefix = "copilot@"
)

func New(config core.Config) (*Agent, error) {
	token, ok := config.Token.(*core.CopilotToken)
	if !ok || token == nil {
		return nil, fmt.Errorf("copilot.New: Token is required")
	}

	return &Agent{
		httpClient: core.NewHTTPClient(),
		model:      config.Model,
		Token:      token,
		efforts:    config.Efforts,
		endpoints:  config.Endpoints,
	}, nil
}

func (a *Agent) Name() string {
	return Prefix + a.model
}

func (a *Agent) authHeader(ctx context.Context) (string, error) {
	refresh, err := oauthCopilot.EnsureFreshSession(ctx, a.Token, a.Refresh)
	if err != nil {
		return "", fmt.Errorf("oauth.EnsureFreshSession: %w", err)
	}
	a.Refresh = refresh
	return "Bearer " + refresh.Token, nil
}
