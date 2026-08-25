package cloudflare

import (
	"fmt"
	"net/http"

	"github.com/pardnchiu/go-llm-router/core"
)

type Agent struct {
	httpClient *http.Client
	model      string
	apiKey     string
	accountID  string
	gatewayID  string
}

const (
	Prefix = "cloudflare@"
)

func New(config core.Config) (*Agent, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("cloudflare.New: APIKey is required")
	}
	if config.AccountID == "" {
		return nil, fmt.Errorf("cloudflare.New: AccountID is required")
	}

	gatewayID := config.GatewayID
	if gatewayID == "" {
		gatewayID = "default"
	}

	return &Agent{
		httpClient: core.NewHTTPClient(),
		model:      config.Model,
		apiKey:     config.APIKey,
		accountID:  config.AccountID,
		gatewayID:  gatewayID,
	}, nil
}

func (a *Agent) Name() string {
	return Prefix + a.model
}
