# Getting Started

Install `go-llm-router`, resolve an Agent from a `provider@model` name, and send a first request.

## Requirements

- Go 1.25 or newer
- `github.com/pardnchiu/go-pkg` v0.13.14, the only direct dependency, pulled in by `go get`
- A credential for the provider you target: an API key, an OAuth token for Copilot, Codex, or Grok, or a logged-in `claude` CLI on `PATH` for `claude-code`
- Token storage through `core/oauth` uses the go-pkg keychain: macOS Keychain on darwin, `secret-tool` elsewhere with a `.secrets` file fallback

## Install

```bash
go get github.com/pardnchiu/go-llm-router
```

Build from source:

```bash
git clone https://github.com/pardnchiu/go-llm-router.git
cd go-llm-router
go build ./...
```

## Send a first request

`Send` always takes five arguments: a context, messages, tool definitions (`nil` is allowed), a `Reasoning` level, and a `Mode`. Every provider shares this signature, so switching models never changes the request loop.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/pardnchiu/go-llm-router/core"
	"github.com/pardnchiu/go-llm-router/core/router"
)

func main() {
	agent, err := router.New(router.Config{
		Name:   "claude@claude-opus-5",
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
	})
	if err != nil {
		log.Fatalf("router.New: %v", err)
	}

	out, code, err := agent.Send(context.Background(),
		[]llmrouter.Message{{Role: "user", Content: "Summarize Go generics in one sentence."}},
		nil, llmrouter.ReasoningDefault, llmrouter.ModeDefault)
	if err != nil {
		log.Fatalf("Send: http %d: %v", code, err)
	}
	if len(out.Choices) == 0 {
		log.Fatal("no choices returned")
	}

	fmt.Println(out.Choices[0].Message.Content)
	fmt.Printf("input=%d output=%d cache_read=%d\n",
		out.Usage.Input, out.Usage.Output, out.Usage.CacheRead)
}
```

Errors carry the upstream HTTP status alongside them, so a retry policy can branch on `code` instead of parsing error strings.

## Next steps

- [Model Routing](/core-concepts-routing): switch providers by changing `Name`
- [Streaming](/core-concepts-streaming): receive the answer as events
- [Model Filters](/core-concepts-model-filters): list the models a provider offers
- [Provider Credentials](/configuration): what each provider needs
