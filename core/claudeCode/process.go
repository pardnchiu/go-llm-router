package claudeCode

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"

	llmrouter "github.com/pardnchiu/go-llm-router/core"
	go_pkg_filesystem "github.com/pardnchiu/go-pkg/filesystem"
	go_pkg_utils "github.com/pardnchiu/go-pkg/utils"
)

const (
	basePrompt  = "You are an agent. Follow the <system_prompt> in the first message as your system prompt."
	plainPrompt = basePrompt + " Reply in plain text."
	toolPrompt  = basePrompt + `
<tools> lists tool names, with the full JSON Schema only for find_tools. Before calling any other tool, fetch its schema with <tool_call name="find_tools">{"query": "select:NAME[,NAME]"}</tool_call>, or by keywords if no name fits; once fetched, call it directly. A new <tools> block replaces the list. Later turns arrive as <user>, <assistant>, <system> and <tool_result> blocks.
Call a tool with one block per call: <tool_call name="TOOL_NAME">{"arg": "value"}</tool_call>, the body a single JSON object of its parameters. Put independent calls in one reply, then stop and wait; results return as <tool_result> blocks in the next message. Text outside <tool_call> blocks is shown to the user; the final answer has none.`
)

const (
	idleTimeout     = 15 * time.Minute
	reapInterval    = time.Minute
	cacheTTLShort   = "5m"
	cacheTTLLong    = "1h"
	placeholderText = "Processing"
	stderrLimit     = 8 << 10
	maxOutputLine   = 64 << 20
)

var (
	poolMu    sync.Mutex
	pool      = map[string]*process{}
	reapStart sync.Once
)

type process struct {
	mu         sync.Mutex
	cmd        *exec.Cmd
	stdin      io.WriteCloser
	stdout     *io.PipeReader
	lines      *bufio.Scanner
	stderr     *limitedBuffer
	exited     chan struct{}
	loaded     bool
	id         string
	cacheTTL   string
	spec       string
	tools      string
	sent       []string
	lastAnswer string
	lastUse    time.Time
}

type limitedBuffer struct {
	mu  sync.Mutex
	buf strings.Builder
}

func (b *limitedBuffer) Write(raw []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if room := stderrLimit - b.buf.Len(); room > 0 {
		b.buf.Write(raw[:min(len(raw), room)])
	}
	return len(raw), nil
}

func (b *limitedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return strings.TrimSpace(b.buf.String())
}

type resultLine struct {
	Type           string `json:"type"`
	Subtype        string `json:"subtype"`
	IsError        bool   `json:"is_error"`
	Result         string `json:"result"`
	APIErrorStatus int    `json:"api_error_status"`
	Usage          struct {
		InputTokens              int `json:"input_tokens"`
		OutputTokens             int `json:"output_tokens"`
		CacheReadInputTokens     int `json:"cache_read_input_tokens"`
		CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	} `json:"usage"`
}

var (
	toolCallOpenPattern  = regexp.MustCompile(`<(?:tool_call|(?:[a-z]+:)?invoke) name="([^"]+)">`)
	toolCallClosePattern = regexp.MustCompile(`</(?:tool_call|(?:[a-z]+:)?invoke)>`)
	parameterPattern     = regexp.MustCompile(`(?s)<(?:[a-z]+:)?parameter name="([^"]+)">(.*?)</(?:[a-z]+:)?parameter>`)
	emptyTagsPattern     = regexp.MustCompile(`^(?:\s*<[A-Za-z_:]+>\s*</[A-Za-z_:]+>\s*)+$`)
)

type toolCallBlock struct {
	name string
	body string
}

func codeFenceRanges(text string) [][2]int {
	var list [][2]int
	fenceStart, fenceChar, fenceLen := -1, byte(0), 0
	for offset := 0; offset < len(text); {
		end := strings.IndexByte(text[offset:], '\n')
		next := len(text)
		if end >= 0 {
			next = offset + end + 1
		}
		line := strings.TrimLeft(text[offset:next], " ")
		if len(text[offset:next])-len(line) <= 3 && line != "" && (line[0] == '`' || line[0] == '~') {
			n := len(line) - len(strings.TrimLeft(line, line[:1]))
			switch {
			case fenceStart < 0 && n >= 3:
				fenceStart, fenceChar, fenceLen = offset, line[0], n
			case fenceStart >= 0 && line[0] == fenceChar && n >= fenceLen && strings.TrimSpace(line[n:]) == "":
				list = append(list, [2]int{fenceStart, next})
				fenceStart = -1
			}
		}
		offset = next
	}
	if fenceStart >= 0 {
		list = append(list, [2]int{fenceStart, len(text)})
	}
	return list
}

func splitToolCalls(text string) (string, []toolCallBlock) {
	var rest strings.Builder
	var list []toolCallBlock
	fences := codeFenceRanges(text)
	cursor, search := 0, 0
	for search < len(text) {
		open := toolCallOpenPattern.FindStringSubmatchIndex(text[search:])
		if open == nil {
			break
		}
		start := search + open[0]
		bodyStart := search + open[1]
		name := text[search+open[2] : search+open[3]]

		if i := slices.IndexFunc(fences, func(r [2]int) bool { return start >= r[0] && start < r[1] }); i >= 0 {
			search = fences[i][1]
			continue
		}

		body, end, ok := jsonToolCallBody(text, bodyStart)
		if !ok {
			closing := toolCallClosePattern.FindStringIndex(text[bodyStart:])
			if closing == nil {
				break
			}
			body = text[bodyStart : bodyStart+closing[0]]
			end = bodyStart + closing[1]
		}
		rest.WriteString(text[cursor:start])
		list = append(list, toolCallBlock{name: name, body: body})
		cursor, search = end, end
	}
	rest.WriteString(text[cursor:])
	return rest.String(), list
}

func jsonToolCallBody(text string, bodyStart int) (string, int, bool) {
	trimmed := strings.TrimLeft(text[bodyStart:], " \t\r\n")
	if !strings.HasPrefix(trimmed, "{") {
		return "", 0, false
	}
	jsonStart := len(text) - len(trimmed)
	decoder := json.NewDecoder(strings.NewReader(trimmed))
	var raw json.RawMessage
	if decoder.Decode(&raw) != nil {
		return "", 0, false
	}
	jsonEnd := jsonStart + int(decoder.InputOffset())
	after := strings.TrimLeft(text[jsonEnd:], " \t\r\n")
	closing := toolCallClosePattern.FindStringIndex(after)
	if closing == nil || closing[0] != 0 {
		return "", 0, false
	}
	return text[jsonStart:jsonEnd], len(text) - len(after) + closing[1], true
}

func invokeArguments(body string) string {
	dic := map[string]any{}
	for _, m := range parameterPattern.FindAllStringSubmatch(body, -1) {
		value := strings.TrimSpace(m[2])
		var parsed any
		if json.Unmarshal([]byte(value), &parsed) == nil {
			dic[m[1]] = parsed
			continue
		}
		dic[m[1]] = value
	}
	raw, err := json.Marshal(dic)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

type processState struct {
	PID        int      `json:"pid"`
	ID         string   `json:"id"`
	Spec       string   `json:"spec"`
	Tools      string   `json:"tools"`
	Sent       []string `json:"sent"`
	LastAnswer string   `json:"last_answer"`
}

var stateFileName = strings.NewReplacer("@", "_", "/", "_", "|", "_")

func statePath(stateDir, sessionID, name string) string {
	if stateDir == "" {
		return ""
	}
	return filepath.Join(stateDir, sessionID, ".claude_code", stateFileName.Replace(name)+".json")
}

func (p *process) restore(path string) {
	state, err := go_pkg_filesystem.ReadJSON[processState](path)
	if err != nil || state.ID == "" {
		return
	}
	if state.PID != os.Getpid() && pidAlive(state.PID) {
		return
	}
	p.id, p.spec, p.tools, p.sent, p.lastAnswer = state.ID, state.Spec, state.Tools, state.Sent, state.LastAnswer
}

func (p *process) save(path string) {
	err := go_pkg_filesystem.WriteJSON(path, processState{
		PID:        os.Getpid(),
		ID:         p.id,
		Spec:       p.spec,
		Tools:      p.tools,
		Sent:       p.sent,
		LastAnswer: p.lastAnswer,
	}, false)
	if err != nil {
		slog.Debug("go_pkg_filesystem.WriteJSON",
			slog.String("path", path),
			slog.String("error", err.Error()))
	}
}

func cacheTTLOf(ctx context.Context) string {
	if llmrouter.CacheTTL(ctx) == cacheTTLLong {
		return cacheTTLLong
	}
	return cacheTTLShort
}

func pidAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}

func acquire(key string) *process {
	reapStart.Do(func() { go reap() })

	for {
		poolMu.Lock()
		p, ok := pool[key]
		if !ok {
			p = &process{}
			pool[key] = p
		}
		poolMu.Unlock()

		p.mu.Lock()
		poolMu.Lock()
		current := pool[key] == p
		poolMu.Unlock()
		if current {
			return p
		}
		p.mu.Unlock()
	}
}

func reap() {
	ticker := time.NewTicker(reapInterval)
	defer ticker.Stop()
	for range ticker.C {
		poolMu.Lock()
		for key, p := range pool {
			if !p.mu.TryLock() {
				continue
			}
			if !p.alive() || time.Since(p.lastUse) > idleTimeout {
				p.stop()
				delete(pool, key)
			}
			p.mu.Unlock()
		}
		poolMu.Unlock()
	}
}

func (p *process) start(model, effort string, withTools bool, sessionID string, resume bool, cacheTTL string) error {
	args := []string{
		"-p",
		"--input-format", "stream-json",
		"--output-format", "stream-json",
		"--verbose",
		"--model", model,
		"--effort", effort,
		"--tools", "",
		"--strict-mcp-config",
		"--safe-mode",
	}
	switch {
	case sessionID == "":
		args = append(args, "--no-session-persistence")
	case resume:
		args = append(args, "--resume", sessionID)
	default:
		args = append(args, "--session-id", sessionID)
	}
	if withTools {
		args = append(args, "--system-prompt", toolPrompt)
	} else {
		args = append(args, "--system-prompt", plainPrompt)
	}

	cmd := exec.Command("claude", args...)
	cmd.Dir = os.TempDir()
	cmd.Env = append(os.Environ(), "CLAUDE_CODE_PROMPT_CACHE_TTL="+cacheTTL)
	if sessionID != "" {
		cmd.Env = append(cmd.Env, "CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1")
	}
	p.cacheTTL = cacheTTL
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("cmd.StdinPipe: %w", err)
	}
	reader, writer := io.Pipe()
	cmd.Stdout = writer
	stderr := &limitedBuffer{}
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("cmd.Start: %w", err)
	}

	exited := make(chan struct{})
	go func() {
		err := cmd.Wait()
		writer.CloseWithError(fmt.Errorf("claude exited: %v", err))
		close(exited)
	}()

	lines := bufio.NewScanner(reader)
	lines.Buffer(make([]byte, 64<<10), maxOutputLine)

	p.cmd, p.stdin, p.stdout, p.lines, p.stderr, p.exited = cmd, stdin, reader, lines, stderr, exited
	return nil
}

func (p *process) alive() bool {
	if p.cmd == nil {
		return false
	}
	select {
	case <-p.exited:
		return false
	default:
		return true
	}
}

func (p *process) stop() {
	if p.cmd == nil {
		return
	}
	_ = p.stdin.Close()
	_ = p.stdout.Close()
	_ = p.cmd.Process.Kill()
	<-p.exited
	p.cmd = nil
}

func (p *process) turn(ctx context.Context, content []map[string]any) (*llmrouter.Output, int, error) {
	raw, err := json.Marshal(map[string]any{
		"type":    "user",
		"message": map[string]any{"role": "user", "content": content},
	})
	if err != nil {
		return nil, 0, fmt.Errorf("json.Marshal: %w", err)
	}
	if _, err := p.stdin.Write(append(raw, '\n')); err != nil {
		return nil, 0, fmt.Errorf("write claude stdin: %w: %s", err, p.stderr.String())
	}

	done := make(chan struct{})
	var line *resultLine
	var readErr error
	go func() {
		defer close(done)
		line, readErr = p.readResult()
	}()

	select {
	case <-done:
	case <-ctx.Done():
		p.stop()
		<-done
		return nil, 0, ctx.Err()
	}

	if readErr != nil {
		return nil, 0, readErr
	}
	if line.IsError {
		return nil, line.APIErrorStatus, fmt.Errorf("claude error (%d): %s", line.APIErrorStatus, go_pkg_utils.TruncateString(line.Result, 1024))
	}
	return buildOutput(line)
}

func (p *process) readResult() (*resultLine, error) {
	for p.lines.Scan() {
		var line resultLine
		if json.Unmarshal(p.lines.Bytes(), &line) != nil || line.Type != "result" {
			continue
		}
		return &line, nil
	}
	if err := p.lines.Err(); err != nil {
		return nil, fmt.Errorf("read claude stdout: %w: %s", err, p.stderr.String())
	}
	return nil, fmt.Errorf("claude exited without a result: %s", p.stderr.String())
}

func buildOutput(line *resultLine) (*llmrouter.Output, int, error) {
	text, blocks := splitToolCalls(line.Result)
	message := llmrouter.Message{
		Role:    "assistant",
		Content: strings.TrimSpace(text),
	}
	placeholder := emptyTagsPattern.MatchString(message.Content.(string))
	if placeholder {
		message.Content = ""
	}
	addCall := func(name, args string) {
		if name = strings.TrimSpace(name); name == "" {
			return
		}
		message.ToolCalls = append(message.ToolCalls, llmrouter.ToolCall{
			ID:       "call_" + strings.ReplaceAll(go_pkg_utils.UUID(), "-", "")[:24],
			Type:     "function",
			Function: llmrouter.ToolCallFunction{Name: name, Arguments: args},
		})
	}
	for _, block := range blocks {
		args := strings.TrimSpace(block.body)
		switch {
		case parameterPattern.MatchString(args):
			args = invokeArguments(args)
		case args == "":
			args = "{}"
		}
		addCall(block.name, args)
	}

	finish := "stop"
	if len(message.ToolCalls) > 0 {
		finish = "tool_calls"
		if placeholder {
			message.Content = placeholderText
		}
	} else if message.Content == "" {
		return nil, 0, fmt.Errorf("claude returned neither an answer nor tool calls")
	}

	u := line.Usage
	return &llmrouter.Output{
		Choices: []llmrouter.OutputChoices{{Message: message, FinishReason: finish}},
		Usage: llmrouter.Usage{
			Input:       u.InputTokens,
			Output:      u.OutputTokens,
			CacheCreate: u.CacheCreationInputTokens,
			CacheRead:   u.CacheReadInputTokens,
		},
	}, 0, nil
}
