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
	"github.com/pardnchiu/go-llm-router/core/mistral"
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

var newFn = map[string]func(config Config) (core.Agent, error){
	"claude": func(config Config) (core.Agent, error) {
		return claude.New(core.Config{Model: strings.TrimPrefix(config.Name, claude.Prefix), APIKey: config.APIKey})
	},
	"openai": func(config Config) (core.Agent, error) {
		return openai.New(core.Config{Model: strings.TrimPrefix(config.Name, openai.Prefix), APIKey: config.APIKey})
	},
	"gemini": func(config Config) (core.Agent, error) {
		return gemini.New(core.Config{Model: strings.TrimPrefix(config.Name, gemini.Prefix), APIKey: config.APIKey})
	},
	"grok": func(config Config) (core.Agent, error) {
		return grok.New(core.Config{Model: strings.TrimPrefix(config.Name, grok.Prefix), APIKey: config.APIKey})
	},
	"deepseek": func(config Config) (core.Agent, error) {
		return deepseek.New(core.Config{Model: strings.TrimPrefix(config.Name, deepseek.Prefix), APIKey: config.APIKey})
	},
	"mistral": func(config Config) (core.Agent, error) {
		return mistral.New(core.Config{Model: strings.TrimPrefix(config.Name, mistral.Prefix), APIKey: config.APIKey})
	},
	"nvidia": func(config Config) (core.Agent, error) {
		return nvidia.New(core.Config{Model: strings.TrimPrefix(config.Name, nvidia.Prefix), APIKey: config.APIKey})
	},
	"openrouter": func(config Config) (core.Agent, error) {
		return openrouter.New(core.Config{Model: strings.TrimPrefix(config.Name, openrouter.Prefix), APIKey: config.APIKey})
	},
	"cloudflare": func(config Config) (core.Agent, error) {
		return cloudflare.New(core.Config{
			Model:     strings.TrimPrefix(config.Name, cloudflare.Prefix),
			APIKey:    config.APIKey,
			AccountID: config.AccountID,
			GatewayID: config.GatewayID,
		})
	},
	"compat": func(config Config) (core.Agent, error) {
		head, model, _ := strings.Cut(config.Name, "@")
		return compat.New(core.Config{
			Model:   model,
			APIKey:  config.APIKey,
			BaseURL: config.BaseURL,
			Prefix:  compatPrefix(head),
		})
	},
	"copilot": func(config Config) (core.Agent, error) {
		return copilot.New(core.Config{Model: strings.TrimPrefix(config.Name, copilot.Prefix), Token: config.Token})
	},
	"codex": func(config Config) (core.Agent, error) {
		return openaicodex.New(core.Config{Model: strings.TrimPrefix(config.Name, openaicodex.Prefix), Token: config.Token})
	},
	"grok-oauth": func(config Config) (core.Agent, error) {
		return grokoauth.New(core.Config{Model: strings.TrimPrefix(config.Name, grokoauth.Prefix), Token: config.Token})
	},
}

func compatPrefix(head string) string {
	_, rest, found := strings.Cut(head, "[")
	if !found {
		return ""
	}
	instance, _, closed := strings.Cut(rest, "]")
	if !closed || instance == "" {
		return ""
	}
	return strings.ToLower(instance) + "@"
}

func New(config Config) (core.Agent, error) {
	providerFull, model, found := strings.Cut(config.Name, "@")
	prov, _, _ := strings.Cut(providerFull, "[")
	fn, ok := newFn[prov]
	if !ok {
		if !found || prov == "" {
			return nil, fmt.Errorf("router.New: unknown provider %q in %q", prov, config.Name)
		}
		config.Name = "compat[" + prov + "]@" + model
		fn = newFn["compat"]
	}
	return fn(config)
}
