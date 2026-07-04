# MCP (Model Context Protocol)

> **Status**: Planned for Phase 9. Not required for RAG.

## What is MCP?

The [Model Context Protocol](https://modelcontextprotocol.io) is a standard protocol
for providing **resources and prompts** to LLMs through a client-server architecture.

```
MCP Server ←→ MCP Client (this library) → LLM
```

MCP is about **providing context** to the LLM — not about the LLM calling functions.
For function calling, see `docs/12-TOOLS.md`.

## Planned MCP Server Integrations

These expose data sources as context for the LLM:

| Server | Resource | Use Case |
|--------|----------|----------|
| **Filesystem** | Files and directories | Let LLM read project files, documentation |
| **GitHub** | Repositories, issues, PRs | Let LLM access code and project management |
| **PostgreSQL** | Database schemas and data | Let LLM query structured data |
| **Slack** | Channel messages and threads | Let LLM access team communication |

## Architecture

The `pkg/mcp/` package will implement an MCP client that:

1. Connects to MCP servers (stdio, HTTP/SSE transports)
2. Discovers available resources and prompts
3. Makes resources available to the LLM via `ChatCompleter`

```go
// Future API sketch
type MCPClient interface {
    Connect(ctx context.Context, server MCPConfig) error
    ListResources(ctx context.Context) ([]Resource, error)
    ReadResource(ctx context.Context, uri string) ([]byte, error)
}
```

## Deferred

MCP is deferred until after the core RAG pipeline (Loader → Embedding → Vector Store → LLM)
is stable. The RAG pipeline does not depend on MCP.
