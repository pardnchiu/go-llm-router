# Claude Code CLI

`claude-code@` 供應商如何把本機安裝的 Claude Code CLI 當作模型後端，以及如何讓每段對話各自維持一個 CLI 程序。

## 前置需求

`claude` 執行檔必須在 `PATH` 中且已登入（先執行一次 `claude`）。此供應商須明確啟用：`EnableClaude` 不為 `true` 時 `router.New` 直接失敗。

```go
agent, err := router.New(router.Config{
	Name:         "claude-code@sonnet",
	EnableClaude: true,
	StateDir:     filepath.Join(home, ".myapp", "sessions"),
})
```

| 欄位 | 作用 |
|---|---|
| `Name` | `claude-code@<model>`；model 原樣傳給 `claude --model` |
| `EnableClaude` | 必填；`false` 時回傳 `claudeCode.New: EnableClaude is required` |
| `StateDir` | 選填；啟用程序重啟後接續對話（見下文） |

`claudeCode.CheckBinary()` 會在執行前回報 CLI 不存在，`claudeCode.Is(name)` 判斷 `provider@model` 名稱是否屬於此供應商。此 Agent 只實作 `Send`：沒有 `SendStream`、`ReasoningLimits`、圖片與語音能力，`Mode` 參數也會被忽略。

## 每段對話一個程序

每次請求都以 stream-JSON 模式執行 `claude -p`，關閉內建工具、MCP server 與專案設定（`--tools ""`、`--strict-mcp-config`、`--safe-mode`），工作目錄為系統暫存目錄。開頭的 `system` 訊息合併後，以 `<system_prompt>` 區塊放在第一輪 user 訊息中送出。

| context | 行為 |
|---|---|
| 未設定 session ID | 每次 `Send` 啟動新程序並帶 `--no-session-persistence`；送出完整歷史，結束後程序退出 |
| 已設定 `WithSessionID` | 每個 session、模型、system prompt 與 `toolDefs` 是否為空的組合對應一個常駐程序，之後只送出上一輪之後新增的訊息 |

只要呼叫端送來的歷史仍以已送出的內容開頭，程序就會沿用。若呼叫端改寫了較早的歷史但保留最新一則回答，只送出該回答之後的訊息。推理等級改變或歷史分岔會以完整歷史重啟程序；system prompt 改變、或在有工具與無工具之間切換，則改用另一個程序，該程序以完整歷史開始。閒置 15 分鐘的程序會被停止。

## Prompt 快取與接續

`llmrouter.WithCacheTTL(ctx, "1h")` 會讓程序帶上 `CLAUDE_CODE_PROMPT_CACHE_TTL=1h`；其他值或未設定時使用 `5m`。TTL 改變會重啟程序。

同時具備 session ID、`1h` TTL 與非空的 `StateDir` 時，對話會被保存：程序取得一個 Claude Code session ID，狀態寫入 `<StateDir>/<sessionID>/.claude_code/<slot>.json`。程式重啟後，下一次呼叫以 `--resume` 接續該 session，不必重送歷史；接續的這一輪失敗時，會以完整歷史重新開始一次。屬於其他仍存活程序的狀態檔會被忽略，任一輪失敗則刪除該檔。

## 推理與輸出對映

| `Reasoning` | `--effort` |
|---|---|
| `ReasoningMedium` | `medium` |
| `ReasoningHigh` | `high` |
| `ReasoningXHigh` | `xhigh` |
| `ReasoningMax` | `max` |
| 其他等級 | `low` |

CLI 的 `result` 行轉成單一 `Output` choice：`input_tokens`、`output_tokens`、`cache_read_input_tokens`、`cache_creation_input_tokens` 分別填入 `Usage.Input`、`Output`、`CacheRead`、`CacheCreate`。`is_error` 的結果以其 `api_error_status` 作為狀態碼回傳；成功的一輪回傳狀態碼 `0`。context 取消時會終止程序並回傳 `ctx.Err()`。圖片 part 為 base64 `data:image/...` URL 或 `http(s)` URL 時以圖片區塊轉送，其他圖片 URL 則換成文字 `[image omitted: unsupported image URL]`。

## 相關頁面

- [Claude Code 工具呼叫](/zh/core-concepts-claude-code-tools)：沒有原生工具支援時工具如何運作
- [對話快取鍵](/zh/core-concepts-sessions)：`WithSessionID` 與 `WithCacheTTL`
- [餘額與配額](/zh/configuration-quota)：`claudeCode.Usage` 訂閱額度
