# Claude Code 工具呼叫

CLI 執行時關閉了原生工具，因此 `claude-code@` 把 `toolDefs` 轉成文字協定，再從 CLI 回覆中解析出工具呼叫。

## 模型看到的內容

`toolDefs` 非空時，程序以工具版 system prompt 啟動，第一輪帶上 `<tools>` 區塊：

```text
<tools>
names: read_file, search_web, find_tools
find_tools: {"name":"find_tools","description":"...","parameters":{...}}
</tools>
```

只列出工具名稱。唯一例外是名為 `find_tools` 的工具：它的完整定義會一併放入，prompt 要求模型在呼叫其他工具前，先以 `{"query": "select:NAME[,NAME]"}`（或關鍵字）呼叫 `find_tools` 取得 schema。傳入較多工具時，請宣告一個會回傳所請求 schema 的 `find_tools` 工具。工具清單在兩輪之間改變時，會送出新的 `<tools>` 區塊取代舊的。

之後的輪次以 `<user>`、`<assistant>`、`<system>`、`<tool_result name="...">` 區塊送出；先前 assistant 的工具呼叫以 `<tool_call>` 行重播。

## 可接受的呼叫格式

prompt 要求模型以下列格式回覆：

```text
<tool_call name="read_file">{"path": "README.md"}</tool_call>
```

解析器也接受模型自行產生的其他形狀：

| 形狀 | 參數 |
|---|---|
| `<tool_call name="X">{json}</tool_call>` | JSON body 原樣使用；body 內含結尾標籤時仍完整讀取 |
| `<invoke name="X">` 搭配 `<parameter name="k">v</parameter>` 子元素 | 重組為 JSON 物件；每個值能以 JSON 解析就解析，否則保留為字串 |
| `<invoke_tool name="X">` 搭配參數 | 同 `invoke` |
| `<invoke_tool>` 搭配 `name` 與 `arguments` 參數 | 名稱與參數由這兩個參數讀取 |

帶命名空間前綴的標籤（`<ns:invoke>`）與 `</parameter_name>` 結尾標籤同樣接受。上述任一標籤都可帶 JSON body 或 `<parameter>` 子元素。

## 不算呼叫的情況

| 情況 | 結果 |
|---|---|
| 標籤位於 fenced code block 或 inline code 內 | 保留為文字 |
| `<tool_call name="x">{}</tool_call>` | 捨棄（佔位，不是呼叫） |
| 名稱為空的呼叫 | 捨棄 |
| body 為空 | 參數設為 `{}` |
| 標籤沒有結尾標籤 | 該標籤及其後所有文字保留為文字 |

## 產出的 Output

每個解析出的呼叫成為一個 `ToolCall`，`Type` 為 `"function"`，ID 為 `call_` 加 24 個十六進位字元，`FinishReason` 為 `tool_calls`。呼叫區塊以外的文字保留在 `Content`；至少解析出一個呼叫且該文字只剩空的標籤對時，`Content` 設為 `Processing`。回覆既無文字也無呼叫時回傳錯誤 `claude returned neither an answer nor tool calls`。

## 相關頁面

- [Claude Code CLI](/zh/core-concepts-claude-code)：前置需求與程序生命週期
- [工具呼叫](/zh/core-concepts-tool-calls)：共用的工具與結果形狀
