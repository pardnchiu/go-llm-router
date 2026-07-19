package cloudflare

import (
	"github.com/pardnchiu/go-llm-router/core"
)

type response struct {
	Result  core.Output `json:"result"`
	Success bool        `json:"success"`
	Errors  []struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"errors"`
}
