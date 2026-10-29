# Claude Code Tool Calls

How `claude-code@` turns `toolDefs` into a text protocol and parses tool calls back out of the CLI's reply, since the CLI runs with its native tools disabled.

## What the model sees

When `toolDefs` is non-empty, the process starts with a tool system prompt and the first turn carries a `<tools>` block:

```text
<tools>
names: read_file, search_web, find_tools
find_tools: {"name":"find_tools","description":"...","parameters":{...}}
</tools>
```

Only tool names are listed. The one exception is a tool named `find_tools`: its full definition is included, and the prompt tells the model to call `find_tools` with `{"query": "select:NAME[,NAME]"}` (or keywords) to fetch a schema before calling any other tool. Declare a `find_tools` tool that returns the requested schemas when you pass more than a handful of tools. When the tool list changes between turns, a new `<tools>` block is sent and replaces the old one.

Later turns arrive as `<user>`, `<assistant>`, `<system>`, and `<tool_result name="...">` blocks; earlier assistant tool calls are replayed as `<tool_call>` lines.

## Call formats accepted

The model is told to answer with:

```text
<tool_call name="read_file">{"path": "README.md"}</tool_call>
```

The parser also accepts the shapes models produce on their own:

| Shape | Arguments |
|---|---|
| `<tool_call name="X">{json}</tool_call>` | the JSON body as-is; a body that contains closing tags is still read whole |
| `<invoke name="X">` with `<parameter name="k">v</parameter>` children | rebuilt as a JSON object; each value is JSON-decoded when possible, otherwise kept as a string |
| `<invoke_tool name="X">` with parameters | same as `invoke` |
| `<invoke_tool>` with `name` and `arguments` parameters | name and arguments read from those parameters |

Namespace-prefixed tags (`<ns:invoke>`) and a `</parameter_name>` closing tag are accepted too. Any of these tags may carry either a JSON body or `<parameter>` children.

## What is not a call

| Case | Result |
|---|---|
| A tag inside a fenced code block or inline code | kept as text |
| `<tool_call name="x">{}</tool_call>` | dropped (a placeholder, not a call) |
| A call with an empty name | dropped |
| An empty body | arguments become `{}` |
| A tag with no closing tag | it and all text after it kept as text |

## Resulting Output

Each parsed call becomes a `ToolCall` with `Type: "function"` and an ID `call_` + 24 hex characters, and `FinishReason` is `tool_calls`. Text outside the call blocks stays in `Content`; when at least one call was parsed and that text is only empty tag pairs, `Content` is set to `Processing`. A reply with neither text nor calls returns the error `claude returned neither an answer nor tool calls`.

## Related

- [Claude Code CLI](/core-concepts-claude-code): requirements and process lifecycle
- [Tool Calls](/core-concepts-tool-calls): the shared tool and result shape
