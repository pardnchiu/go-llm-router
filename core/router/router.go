package router

import (
	"fmt"
	"strings"

	"github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/claude"
	"github.com/pardnchiu/go-llm-router/core/cloudflare"
	"github.com/pardnchiu/go-llm-router/core/compat"
	"github.com/pardnchiu/go-llm-router/core/copilot"
	"github.com/pardnchiu/go-llm-router/core/deepseek"
	"github.com/pardnchiu/go-llm-router/core/gemini"
	"github.com/pardnchiu/go-llm-router/core/grok"
	grokoauth "github.com/pardnchiu/go-llm-router/core/grokOauth"
	"github.com/pardnchiu/go-llm-router/core/nvidia"
	openrouter "github.com/pardnchiu/go-llm-router/core/openRouter"
	"github.com/pardnchiu/go-llm-router/core/openai"
	openaicodex "github.com/pardnchiu/go-llm-router/core/openaiCodex"
)

type Config struct {
	Name    string
	APIKey  string
	Token   any
	BaseURL string

	AccountID string
	GatewayID string
}

var newFn = map[string]func(config Config) (provider.Agent, error){
	"claude": func(config Config) (provider.Agent, error) {
		return claude.New(provider.Config{Model: strings.TrimPrefix(config.Name, claude.Prefix), APIKey: config.APIKey})
	},
	"openai": func(config Config) (provider.Agent, error) {
		return openai.New(provider.Config{Model: strings.TrimPrefix(config.Name, openai.Prefix), APIKey: config.APIKey})
	},
	"gemini": func(config Config) (provider.Agent, error) {
		return gemini.New(provider.Config{Model: strings.TrimPrefix(config.Name, gemini.Prefix), APIKey: config.APIKey})
	},
	"grok": func(config Config) (provider.Agent, error) {
		return grok.New(provider.Config{Model: strings.TrimPrefix(config.Name, grok.Prefix), APIKey: config.APIKey})
	},
	"deepseek": func(config Config) (provider.Agent, error) {
		return deepseek.New(provider.Config{Model: strings.TrimPrefix(config.Name, deepseek.Prefix), APIKey: config.APIKey})
	},
	"nvidia": func(config Config) (provider.Agent, error) {
		return nvidia.New(provider.Config{Model: strings.TrimPrefix(config.Name, nvidia.Prefix), APIKey: config.APIKey})
	},
	"openrouter": func(config Config) (provider.Agent, error) {
		return openrouter.New(provider.Config{Model: strings.TrimPrefix(config.Name, openrouter.Prefix), APIKey: config.APIKey})
	},
	"cloudflare": func(config Config) (provider.Agent, error) {
		return cloudflare.New(provider.Config{
			Model:     strings.TrimPrefix(config.Name, cloudflare.Prefix),
			APIKey:    config.APIKey,
			AccountID: config.AccountID,
			GatewayID: config.GatewayID,
		})
	},
	"compat": func(config Config) (provider.Agent, error) {
		_, model, _ := strings.Cut(config.Name, "@")
		return compat.New(provider.Config{
			Model:   model,
			APIKey:  config.APIKey,
			BaseURL: config.BaseURL,
		})
	},
	"copilot": func(config Config) (provider.Agent, error) {
		return copilot.New(provider.Config{Model: strings.TrimPrefix(config.Name, copilot.Prefix), Token: config.Token})
	},
	"codex": func(config Config) (provider.Agent, error) {
		return openaicodex.New(provider.Config{Model: strings.TrimPrefix(config.Name, openaicodex.Prefix), Token: config.Token})
	},
	"grok-oauth": func(config Config) (provider.Agent, error) {
		return grokoauth.New(provider.Config{Model: strings.TrimPrefix(config.Name, grokoauth.Prefix), Token: config.Token})
	},
}

func New(config Config) (provider.Agent, error) {
	providerFull, _, _ := strings.Cut(config.Name, "@")
	prov, _, _ := strings.Cut(providerFull, "[")
	fn, ok := newFn[prov]
	if !ok {
		return nil, fmt.Errorf("router.New: unknown provider %q in %q", prov, config.Name)
	}
	return fn(config)
}
