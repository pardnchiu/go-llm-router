package llmrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"
)

const (
	StreamProbeInterval    = 30 * time.Second
	StreamRecheckInterval  = 10 * time.Second
	StreamMaxProbeFailures = 3
	StreamRepeatWindow     = 16 << 10
	StreamRepeatMaxPeriod  = 4 << 10
	StreamRepeatMaxBytes   = 4
	StreamRepeatCheckBytes = 512
	streamRepeatQuoteBytes = 128
)

type idleBody struct {
	io.ReadCloser
	label       string
	reads       atomic.Int64
	frames      atomic.Int64
	content     atomic.Int64
	seenReads   int64
	seenFrames  int64
	seenContent int64
	failures    int
	contentless bool
	decodeFails int
	closed      bool
	mu          sync.Mutex
	timer       *time.Timer
	inRead      atomic.Bool
	fired       atomic.Bool
	recent      []byte
	unchecked   int
	repeated    bool
}

func WatchIdle(body io.ReadCloser, label string) io.ReadCloser {
	b := &idleBody{ReadCloser: body, label: label}
	b.mu.Lock()
	b.timer = time.AfterFunc(StreamProbeInterval, b.probe)
	b.mu.Unlock()
	return b
}

func (b *idleBody) probe() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	reads, frames, content := b.reads.Load(), b.frames.Load(), b.content.Load()
	gotData := reads != b.seenReads
	contentless := frames != b.seenFrames && content == b.seenContent
	b.seenReads, b.seenFrames, b.seenContent = reads, frames, content
	if (gotData && !contentless) || !b.inRead.Load() {
		b.failures = 0
		b.timer.Reset(StreamProbeInterval)
		return
	}
	b.failures++
	if b.failures < StreamMaxProbeFailures {
		b.timer.Reset(StreamRecheckInterval)
		return
	}
	b.contentless = contentless
	b.fired.Store(true)
	b.ReadCloser.Close()
}

func (b *idleBody) Read(p []byte) (int, error) {
	b.inRead.Store(true)
	n, err := b.ReadCloser.Read(p)
	b.inRead.Store(false)
	if b.repeated {
		quote := b.recent[len(b.recent)-streamRepeatQuoteBytes:]
		return n, fmt.Errorf("[Timeout] %s stream: degenerate repetition in last %d bytes: %q: %w", b.label, StreamRepeatWindow, bytes.ToValidUTF8(quote, nil), context.DeadlineExceeded)
	}
	if b.fired.Load() {
		silent := StreamProbeInterval + StreamRecheckInterval*(StreamMaxProbeFailures-1)
		if b.contentless {
			return n, fmt.Errorf("[Timeout] %s stream: data without parsable content for %s: %w", b.label, silent, context.DeadlineExceeded)
		}
		return n, fmt.Errorf("[Timeout] %s stream: no data for %s: %w", b.label, silent, context.DeadlineExceeded)
	}
	if n == 0 {
		return n, err
	}
	b.reads.Add(1)
	if !keepAliveOnly(p[:n]) {
		b.frames.Add(1)
	}
	return n, err
}

func TouchIdle(body io.Reader, deltas ...string) {
	b, ok := body.(*idleBody)
	if !ok {
		return
	}
	b.content.Add(1)
	if b.repeated {
		return
	}
	grown := len(b.recent)
	for _, delta := range deltas {
		b.recent = append(b.recent, delta...)
	}
	b.unchecked += len(b.recent) - grown
	if b.unchecked < StreamRepeatCheckBytes || len(b.recent) < StreamRepeatWindow {
		return
	}
	b.unchecked = 0
	b.recent = b.recent[len(b.recent)-StreamRepeatWindow:]
	b.repeated = repetitive(b.recent)
	if b.repeated {
		b.ReadCloser.Close()
	}
}

func repetitive(window []byte) bool {
	var seen [256]bool
	distinct := 0
	for _, c := range window {
		if !seen[c] {
			seen[c] = true
			distinct++
		}
	}
	if distinct <= StreamRepeatMaxBytes {
		return true
	}
	for period := 1; period <= StreamRepeatMaxPeriod; period++ {
		if bytes.Equal(window[period:], window[:len(window)-period]) {
			return true
		}
	}
	return false
}

func DecodeFrame(body io.Reader, label, data string, v any) (bool, error) {
	err := json.Unmarshal([]byte(data), v)
	b, ok := body.(*idleBody)
	if !ok {
		if err != nil {
			return false, fmt.Errorf("[Upstream] %s stream decode: %w: %s", label, err, TruncateFrame(data))
		}
		return true, nil
	}
	if err == nil {
		b.decodeFails = 0
		return true, nil
	}
	b.decodeFails++
	if b.decodeFails < StreamMaxProbeFailures {
		return false, nil
	}
	return false, fmt.Errorf("[Upstream] %s stream decode: %d consecutive frames: %w: %s", label, b.decodeFails, err, TruncateFrame(data))
}

func (b *idleBody) Close() error {
	b.mu.Lock()
	b.closed = true
	b.timer.Stop()
	b.mu.Unlock()
	return b.ReadCloser.Close()
}

func keepAliveOnly(chunk []byte) bool {
	for line := range bytes.Lines(chunk) {
		line = bytes.TrimSpace(line)
		if len(line) > 0 && line[0] != ':' {
			return false
		}
	}
	return true
}
