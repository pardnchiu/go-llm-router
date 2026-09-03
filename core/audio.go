package core

import (
	"context"
	"encoding/binary"
	"strconv"
	"strings"
)

type STTOptions struct {
	Prompt   string
	Language string
	MimeType string
}

type STTResult struct {
	Text string
}

type STTAgent interface {
	Transcribe(ctx context.Context, audio []byte, opts STTOptions) (*STTResult, error)
}

type TTSOptions struct {
	Voice string
}

type TTSResult struct {
	Audio    []byte
	MimeType string
}

type TTSAgent interface {
	Speak(ctx context.Context, text string, opts TTSOptions) (*TTSResult, error)
}

func PCMRate(mime string) int {
	for part := range strings.SplitSeq(mime, ";") {
		key, value, found := strings.Cut(strings.TrimSpace(part), "=")
		if !found || strings.ToLower(key) != "rate" {
			continue
		}
		if rate, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && rate > 0 {
			return rate
		}
	}
	return 24000
}

func WrapPCM16(pcm []byte, rate int) []byte {
	const channels, bits = 1, 16
	byteRate := rate * channels * bits / 8

	header := make([]byte, 44)
	copy(header[0:], "RIFF")
	binary.LittleEndian.PutUint32(header[4:], uint32(36+len(pcm)))
	copy(header[8:], "WAVEfmt ")
	binary.LittleEndian.PutUint32(header[16:], 16)
	binary.LittleEndian.PutUint16(header[20:], 1)
	binary.LittleEndian.PutUint16(header[22:], channels)
	binary.LittleEndian.PutUint32(header[24:], uint32(rate))
	binary.LittleEndian.PutUint32(header[28:], uint32(byteRate))
	binary.LittleEndian.PutUint16(header[32:], channels*bits/8)
	binary.LittleEndian.PutUint16(header[34:], bits)
	copy(header[36:], "data")
	binary.LittleEndian.PutUint32(header[40:], uint32(len(pcm)))

	return append(header, pcm...)
}
