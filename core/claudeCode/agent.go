package claudeCode

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
	"time"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	go_pkg_utils "github.com/pardnchiu/go-pkg/utils"
)

var messagePrefixPattern = regexp.MustCompile(`\A\s*(?:sendAt|sender|channelId)\s*:[^\n]*(?:\n|\z)`)

func CheckBinary() error {
	if _, err := exec.LookPath("claude"); err != nil {
		return fmt.Errorf("claude CLI not found; install Claude Code and run `claude` once to log in: %w", err)
	}
	return nil
}

func Is(name string) bool {
	prov, _, _ := strings.Cut(name, "@")
	return prov == llmrouter.PROVIDER_CLAUDE_CODE
}

type Agent struct {
	model    string
	stateDir string
}

func New(config llmrouter.Config) (*Agent, error) {
	if config.Model == "" {
		return nil, fmt.Errorf("claudeCode.New: Model is required")
	}
	if err := CheckBinary(); err != nil {
		return nil, err
	}
	if !config.EnableClaude {
		return nil, fmt.Errorf("claudeCode.New: EnableClaude is required")
	}
	return &Agent{model: config.Model, stateDir: config.StateDir}, nil
}

func (a *Agent) Name() string {
	return llmrouter.PROVIDER_CLAUDE_CODE + "@" + a.model
}

func (a *Agent) Send(ctx context.Context, messages []llmrouter.Message, toolDefs []llmrouter.Tool, reasoning llmrouter.Reasoning, _ llmrouter.Mode) (*llmrouter.Output, int, error) {
	effort := effortOf(reasoning)
	system, rest := splitSystem(messages)
	sessionID := llmrouter.SessionID(ctx)

	if sessionID == "" {
		p := &process{}
		if err := p.start(a.model, effort, len(toolDefs) > 0, "", false, cacheTTLShort); err != nil {
			return nil, 0, err
		}
		defer p.stop()
		return p.turn(ctx, renderInitial(system, toolDefs, rest))
	}

	withTools := len(toolDefs) > 0
	sum := sha256.Sum256([]byte(system))
	slot := a.model + "|" + hex.EncodeToString(sum[:8])
	if !withTools {
		slot += "|plain"
	}
	p := acquire(sessionID + "|" + slot)
	defer p.mu.Unlock()

	spec := specOf(system, effort)
	tools := renderTools(toolDefs)
	list := fingerprints(rest)
	cacheTTL := cacheTTLOf(ctx)
	path := statePath(a.stateDir, sessionID, slot)
	persist := cacheTTL == cacheTTLLong && path != ""
	if p.alive() && p.cacheTTL != cacheTTL {
		p.stop()
	}

	if persist && !p.loaded {
		p.loaded = true
		p.restore(path)
	}

	startFresh := func() ([]map[string]any, error) {
		p.stop()
		p.id = ""
		if persist {
			p.id = go_pkg_utils.UUID()
		}
		if err := p.start(a.model, effort, withTools, p.id, false, cacheTTL); err != nil {
			return nil, err
		}
		p.spec = spec
		p.tools = tools
		p.sent = nil
		return renderInitial(system, toolDefs, rest), nil
	}

	reusable := p.spec == spec && (p.alive() || (persist && p.id != ""))
	resumed := false
	var content []map[string]any
	if reusable && len(list) > len(p.sent) && slices.Equal(list[:len(p.sent)], p.sent) {
		content = renderMessages(rest[len(p.sent):], toolNames(rest))
	} else if i := answerIndex(rest, p.lastAnswer); reusable && i >= 0 {
		content = renderMessages(rest[i+1:], toolNames(rest))
	} else {
		reusable = false
	}
	if !reusable {
		fresh, err := startFresh()
		if err != nil {
			return nil, 0, err
		}
		content = fresh
	} else if !p.alive() {
		p.stop()
		if err := p.start(a.model, effort, withTools, p.id, true, cacheTTL); err != nil {
			return nil, 0, err
		}
		resumed = true
	}
	if tools != p.tools {
		content = append([]map[string]any{{"type": "text", "text": tools + "\n\n"}}, content...)
		p.tools = tools
	}

	out, code, err := p.turn(ctx, content)
	if err != nil && resumed && ctx.Err() == nil {
		fresh, startErr := startFresh()
		if startErr != nil {
			return nil, 0, startErr
		}
		out, code, err = p.turn(ctx, fresh)
	}
	if err != nil {
		p.stop()
		p.id = ""
		_ = os.Remove(path)
		return nil, code, err
	}
	message := out.Choices[0].Message
	p.sent = append(list, fingerprint(message))
	p.lastAnswer = ""
	if len(message.ToolCalls) == 0 {
		p.lastAnswer = answerText(message.Content)
	}
	p.lastUse = time.Now()
	if persist {
		p.save(path)
	}
	return out, code, nil
}

func answerIndex(messages []llmrouter.Message, answer string) int {
	if answer == "" {
		return -1
	}
	for i, m := range slices.Backward(messages) {
		if m.Role != "assistant" || len(m.ToolCalls) > 0 {
			continue
		}
		if i == len(messages)-1 || answerText(m.Content) != answer {
			return -1
		}
		return i
	}
	return -1
}

func answerText(content any) string {
	return strings.TrimSpace(messagePrefixPattern.ReplaceAllString(contentText(content), ""))
}

func effortOf(reasoning llmrouter.Reasoning) string {
	switch reasoning {
	case llmrouter.ReasoningMedium:
		return "medium"
	case llmrouter.ReasoningHigh:
		return "high"
	case llmrouter.ReasoningXHigh:
		return "xhigh"
	case llmrouter.ReasoningMax:
		return "max"
	}
	return "low"
}
