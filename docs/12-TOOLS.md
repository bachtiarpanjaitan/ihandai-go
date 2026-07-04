# Tools (Function Calling)

> **Status**: Core interface defined now. Full execution framework planned for Phase 7 (Agents).

## What are Tools?

Tools are functions that an LLM can request to call. This is distinct from MCP
(Model Context Protocol), which provides resources and context to an LLM.

```
User: "Send an email to the team about the deploy"
  → LLM decides to call send_email tool
    → Framework executes send_email()
      → Result returned to LLM
        → LLM: "Email sent successfully"
```

## Tool Interface (`pkg/tools/`)

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() *JSONSchema
    Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error)
}
```

## Planned Tools

| Tool | Description | Use Case |
|------|-------------|----------|
| **Browser** | Open a URL and extract content | Web search, documentation lookup |
| **Slack** | Send messages to Slack channels | Notifications, reports |
| **Email** | Send emails via SMTP | Alerts, summaries |
| **HTTP Request** | Make arbitrary HTTP calls | API integrations |
| **Code Executor** | Run code in a sandbox | Calculations, data processing |
| **Database Query** | Query databases directly | Data lookups |

## How Tools Integrate with LLM

When a `ChatCompleter` receives a request, the library:

1. Passes available tool definitions alongside the messages
2. If the LLM responds with a tool call, the library executes it
3. The tool's result is sent back to the LLM for the final response

```go
// Future API sketch
ai := ihandai.New(
    ihandai.WithLLM("openai", ...),
    ihandai.WithTools(
        tools.NewBrowser(),
        tools.NewEmail(smtpConfig),
    ),
)

// LLM can now decide to call browser or email as needed
resp, _ := ai.Ask(ctx, "Find the latest docs and email them to the team")
```

## MCP vs Tools

| | MCP | Tools |
|---|---|---|
| **Purpose** | Provide context/resources TO the LLM | Functions the LLM can CALL |
| **Initiative** | Framework connects to servers | LLM decides when to call |
| **Protocol** | MCP standard (modelcontextprotocol.io) | LLM API function calling |
| **Examples** | Filesystem, GitHub, PostgreSQL | Browser, Email, Slack, HTTP |
| **Timeline** | Phase 9 | Phase 7 (with Agents) |
| **Package** | `pkg/mcp/` | `pkg/tools/` |
