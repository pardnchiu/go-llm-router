package ollamacloud

import (
	"fmt"
	"net/http"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

type Agent struct {
	httpClient *http.Client
	model      string
	apiKey     string
}

const (
	Prefix = "ollama-cloud@"
)

func New(config llmrouter.Config) (*Agent, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("ollamacloud.New: APIKey is required")
	}

	return &Agent{
		httpClient: llmrouter.NewHTTPClient(),
		model:      config.Model,
		apiKey:     config.APIKey,
	}, nil
}

func (a *Agent) Name() string {
	return Prefix + a.model
}
