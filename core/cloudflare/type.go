package cloudflare

import (
	llmrouter "github.com/pardnchiu/go-llm-router/core"
)

type response struct {
	Result  llmrouter.Output `json:"result"`
	Success bool             `json:"success"`
	Errors  []struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"errors"`
}
