package gemini

import (
	"fmt"
	"net/http"
	"sync"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

const (
	pingInterval = llmrouter.StreamProbeInterval
	pingTimeout  = llmrouter.StreamRecheckInterval * (llmrouter.StreamMaxProbeFailures - 1)
)

type Agent struct {
	httpClient *http.Client
	model      string
	apiKey     string
	cacheMu    sync.Mutex
	cacheStore map[string]*geminiCacheEntry
	cacheMode  cacheSupport
	thinking   *bool
}

func New(config llmrouter.Config) (*Agent, error) {
	if config.APIKey == "" {
		return nil, fmt.Errorf("gemini.New: APIKey is required")
	}

	return &Agent{
		httpClient: newHTTPClient(),
		model:      config.Model,
		apiKey:     config.APIKey,
		cacheStore: make(map[string]*geminiCacheEntry),
		thinking:   config.Thinking,
	}, nil
}

func (a *Agent) Name() string {
	return llmrouter.PROVIDER_GEMINI + "@" + a.model
}

func newHTTPClient() *http.Client {
	client := llmrouter.NewHTTPClient()
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.ForceAttemptHTTP2 = true
	transport.HTTP2 = &http.HTTP2Config{
		SendPingTimeout: pingInterval,
		PingTimeout:     pingTimeout,
	}
	client.Transport = transport
	return client
}
