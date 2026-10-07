package router

import (
	"fmt"
	"strings"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/claude"
	"github.com/pardnchiu/go-llm-router/core/claudeCode"
	"github.com/pardnchiu/go-llm-router/core/cloudflare"
	"github.com/pardnchiu/go-llm-router/core/compat"
	"github.com/pardnchiu/go-llm-router/core/copilot"
	"github.com/pardnchiu/go-llm-router/core/deepseek"
	"github.com/pardnchiu/go-llm-router/core/gemini"
	"github.com/pardnchiu/go-llm-router/core/grok"
	grokoauth "github.com/pardnchiu/go-llm-router/core/grokOauth"
	"github.com/pardnchiu/go-llm-router/core/mistral"
	"github.com/pardnchiu/go-llm-router/core/nvidia"
	ollamacloud "github.com/pardnchiu/go-llm-router/core/ollamaCloud"
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

	EnableClaude bool
	StateDir     string
}

var newFn = map[string]func(config Config) (llmrouter.Agent, error){
	llmrouter.PROVIDER_CLAUDE: func(config Config) (llmrouter.Agent, error) {
		return claude.New(llmrouter.Config{Model: strings.TrimPrefix(config.Name, llmrouter.PROVIDER_CLAUDE+"@"), APIKey: config.APIKey})
	},
	llmrouter.PROVIDER_CLAUDE_CODE: func(config Config) (llmrouter.Agent, error) {
		return claudeCode.New(llmrouter.Config{
			Model:        strings.TrimPrefix(config.Name, llmrouter.PROVIDER_CLAUDE_CODE+"@"),
			EnableClaude: config.EnableClaude,
			StateDir:     config.StateDir,
		})
	},
	llmrouter.PROVIDER_OPENAI: func(config Config) (llmrouter.Agent, error) {
		return openai.New(llmrouter.Config{Model: strings.TrimPrefix(config.Name, llmrouter.PROVIDER_OPENAI+"@"), APIKey: config.APIKey})
	},
	llmrouter.PROVIDER_GEMINI: func(config Config) (llmrouter.Agent, error) {
		return gemini.New(llmrouter.Config{Model: strings.TrimPrefix(config.Name, llmrouter.PROVIDER_GEMINI+"@"), APIKey: config.APIKey})
	},
	llmrouter.PROVIDER_GROK: func(config Config) (llmrouter.Agent, error) {
		return grok.New(llmrouter.Config{Model: strings.TrimPrefix(config.Name, llmrouter.PROVIDER_GROK+"@"), APIKey: config.APIKey})
	},
	llmrouter.PROVIDER_DEEPSEEK: func(config Config) (llmrouter.Agent, error) {
		return deepseek.New(llmrouter.Config{Model: strings.TrimPrefix(config.Name, llmrouter.PROVIDER_DEEPSEEK+"@"), APIKey: config.APIKey})
	},
	llmrouter.PROVIDER_MISTRAL: func(config Config) (llmrouter.Agent, error) {
		return mistral.New(llmrouter.Config{Model: strings.TrimPrefix(config.Name, llmrouter.PROVIDER_MISTRAL+"@"), APIKey: config.APIKey})
	},
	llmrouter.PROVIDER_NVIDIA: func(config Config) (llmrouter.Agent, error) {
		return nvidia.New(llmrouter.Config{Model: strings.TrimPrefix(config.Name, llmrouter.PROVIDER_NVIDIA+"@"), APIKey: config.APIKey})
	},
	llmrouter.PROVIDER_OLLAMA_CLOUD: func(config Config) (llmrouter.Agent, error) {
		return ollamacloud.New(llmrouter.Config{Model: strings.TrimPrefix(config.Name, llmrouter.PROVIDER_OLLAMA_CLOUD+"@"), APIKey: config.APIKey})
	},
	llmrouter.PROVIDER_OPENROUTER: func(config Config) (llmrouter.Agent, error) {
		return openrouter.New(llmrouter.Config{Model: strings.TrimPrefix(config.Name, llmrouter.PROVIDER_OPENROUTER+"@"), APIKey: config.APIKey})
	},
	llmrouter.PROVIDER_CLOUDFLARE: func(config Config) (llmrouter.Agent, error) {
		return cloudflare.New(llmrouter.Config{
			Model:     strings.TrimPrefix(config.Name, llmrouter.PROVIDER_CLOUDFLARE+"@"),
			APIKey:    config.APIKey,
			AccountID: config.AccountID,
			GatewayID: config.GatewayID,
		})
	},
	llmrouter.PROVIDER_COMPAT: func(config Config) (llmrouter.Agent, error) {
		head, model, _ := strings.Cut(config.Name, "@")
		return compat.New(llmrouter.Config{
			Model:   model,
			APIKey:  config.APIKey,
			BaseURL: config.BaseURL,
			Prefix:  compatPrefix(head),
		})
	},
	llmrouter.PROVIDER_COPILOT: func(config Config) (llmrouter.Agent, error) {
		return copilot.New(llmrouter.Config{Model: strings.TrimPrefix(config.Name, llmrouter.PROVIDER_COPILOT+"@"), Token: config.Token})
	},
	llmrouter.PROVIDER_CODEX: func(config Config) (llmrouter.Agent, error) {
		return openaicodex.New(llmrouter.Config{Model: strings.TrimPrefix(config.Name, llmrouter.PROVIDER_CODEX+"@"), Token: config.Token})
	},
	llmrouter.PROVIDER_GROK_OAUTH: func(config Config) (llmrouter.Agent, error) {
		return grokoauth.New(llmrouter.Config{Model: strings.TrimPrefix(config.Name, llmrouter.PROVIDER_GROK_OAUTH+"@"), Token: config.Token})
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

func New(config Config) (llmrouter.Agent, error) {
	providerFull, model, found := strings.Cut(config.Name, "@")
	prov, _, _ := strings.Cut(providerFull, "[")
	fn, ok := newFn[prov]
	if !ok {
		if !found || prov == "" {
			return nil, fmt.Errorf("router.New: unknown provider %q in %q", prov, config.Name)
		}
		config.Name = llmrouter.PROVIDER_COMPAT + "[" + prov + "]@" + model
		fn = newFn[llmrouter.PROVIDER_COMPAT]
	}
	return fn(config)
}
