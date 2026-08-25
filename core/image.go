package core

import (
	"context"
	"strconv"
	"strings"
)

type ImageOptions struct {
	AspectRatio string // * 1:1/16:9/4:3, xai/gemini native, size with openai
	Size        string // * 1k/2k/4k, xai resolution/gemini imageSize/openai w*h
	Quality     string // * low/medium/high, skip in gemini
	RefImageB64 string
	RefMime     string
}

type ImageResult struct {
	B64      string
	MimeType string
	Revised  string
}

type ImageAgent interface {
	GenerateImage(ctx context.Context, prompt string, opts ImageOptions) (*ImageResult, error)
}

var sizeEdge = map[string]int{"1k": 1024, "2k": 2048, "4k": 4096}

func ImagePixelSize(opts ImageOptions) string {
	edge, ok := sizeEdge[strings.ToLower(opts.Size)]
	if !ok {
		if opts.AspectRatio == "" {
			return ""
		}
		edge = 1024
	}

	w, h, ok := parseRatio(opts.AspectRatio)
	if !ok {
		return strconv.Itoa(edge) + "x" + strconv.Itoa(edge)
	}
	if w >= h {
		return strconv.Itoa(round16(float64(edge)*w/h)) + "x" + strconv.Itoa(edge)
	}
	return strconv.Itoa(edge) + "x" + strconv.Itoa(round16(float64(edge)*h/w))
}

func parseRatio(ratio string) (w, h float64, ok bool) {
	left, right, found := strings.Cut(ratio, ":")
	if !found {
		return 0, 0, false
	}
	w, err := strconv.ParseFloat(strings.TrimSpace(left), 64)
	if err != nil || w <= 0 {
		return 0, 0, false
	}
	h, err = strconv.ParseFloat(strings.TrimSpace(right), 64)
	if err != nil || h <= 0 {
		return 0, 0, false
	}
	return w, h, true
}

func round16(v float64) int {
	n := int((v + 8) / 16)
	return max(n, 1) * 16
}

func DataURI(mime, b64 string) string {
	if mime == "" {
		mime = "image/png"
	}
	return "data:" + mime + ";base64," + b64
}
