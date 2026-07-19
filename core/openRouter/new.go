package openrouter

import (
	"fmt"
	"net/http"

	"github.com/pardnchiu/go-llm-router/core"
)

type Agent struct {
	httpClient *http.Client
	model      string
	apiKey     string
}

const Prefix = "openrouter@"

func New(config core.Config) (*Agent, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("openrouter.New: APIKey is required")
	}

	return &Agent{
		httpClient: core.NewHTTPClient(),
		model:      config.Model,
		apiKey:     config.APIKey,
	}, nil
}

func (a *Agent) Name() string {
	return Prefix + a.model
}
