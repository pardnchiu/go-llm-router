package mistral

import (
	"fmt"
	"net/http"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

type Agent struct {
	httpClient *http.Client
	model      string
	apiKey     string
	thinking   *bool
}

const (
	Prefix = "mistral@"
)

func New(config llmrouter.Config) (*Agent, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("mistral.New: APIKey is required")
	}

	return &Agent{
		httpClient: llmrouter.NewHTTPClient(),
		model:      config.Model,
		apiKey:     config.APIKey,
		thinking:   config.Thinking,
	}, nil
}

func (a *Agent) Name() string {
	return Prefix + a.model
}
