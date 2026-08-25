package compat

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pardnchiu/go-llm-router/core"
)

type Agent struct {
	httpClient *http.Client
	model      string
	baseURL    string
	apiKey     string
	prefix     string
}

const (
	defaultBaseURL = "http://localhost:11434/v1"
	defaultPrefix  = "compat@"
)

func New(config core.Config) (*Agent, error) {
	if config.Model == "" {
		return nil, fmt.Errorf("compat.New: Model is required")
	}

	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	baseURL = strings.TrimRight(baseURL, "/")

	prefix := config.Prefix
	if prefix == "" {
		prefix = defaultPrefix
	}

	return &Agent{
		httpClient: &http.Client{Timeout: 10 * time.Minute},
		model:      config.Model,
		baseURL:    baseURL,
		apiKey:     config.APIKey,
		prefix:     prefix,
	}, nil
}

func (a *Agent) Name() string {
	return a.prefix + a.model
}
