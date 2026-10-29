# 快速開始

安裝 `go-llm-router`，用 `provider@model` 取得 Agent，並送出第一個請求。

## 需求

- Go 1.25 以上
- `github.com/pardnchiu/go-pkg` v0.13.14（唯一直接相依，`go get` 時自動取得）
- 目標供應商的憑證：API 金鑰、Copilot／Codex／Grok 的 OAuth 權杖，或 `claude-code` 所需、位於 `PATH` 且已登入的 `claude` CLI
- `core/oauth` 透過 go-pkg keychain 保存權杖：darwin 用 macOS Keychain，其他平台用 `secret-tool`，失敗時退回 `.secrets` 檔

## 安裝

```bash
go get github.com/pardnchiu/go-llm-router
```

從原始碼建置：

```bash
git clone https://github.com/pardnchiu/go-llm-router.git
cd go-llm-router
go build ./...
```

## 送出第一個請求

`Send` 固定五個參數：context、訊息、工具定義（可傳 `nil`）、`Reasoning` 等級、`Mode`。所有供應商共用這組簽名，換模型不必改請求迴圈。

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
		[]llmrouter.Message{{Role: "user", Content: "用一句話說明 Go 的 generics。"}},
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

錯誤回傳會同時帶上游 HTTP 狀態碼，因此重試策略可以直接依 `code` 判斷，不需要解析錯誤字串。

## 下一步

- [模型路由](/zh/core-concepts-routing)：改 `Name` 就能換供應商
- [串流](/zh/core-concepts-streaming)：以事件形式接收回應
- [模型篩選](/zh/core-concepts-model-filters)：列出供應商提供的模型
- [供應商憑證](/zh/configuration)：各供應商需要的欄位
