package gemini

import (
	"fmt"
	"net/http"
	"sync"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

type Agent struct {
	httpClient *http.Client
	model      string
	apiKey     string
	cacheMu    sync.Mutex
	cacheStore map[string]*geminiCacheEntry
	thinking   *bool
}

const (
	Prefix = "gemini@"
)

func New(config llmrouter.Config) (*Agent, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("gemini.New: APIKey is required")
	}

	return &Agent{
		httpClient: llmrouter.NewHTTPClient(),
		model:      config.Model,
		apiKey:     config.APIKey,
		cacheStore: make(map[string]*geminiCacheEntry),
		thinking:   config.Thinking,
	}, nil
}

func (a *Agent) Name() string {
	return Prefix + a.model
}
