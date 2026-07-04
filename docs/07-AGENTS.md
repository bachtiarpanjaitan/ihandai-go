# Agent Framework

> **Status**: Planned for Phase 7. Depends on Phase 5 (RAG) and Phase 6 (Memory).

## What is an Agent?

An agent is an autonomous loop where an LLM plans, executes tools, observes results,
reflects, and decides the next action — rather than responding in a single pass.

```
User goal → Planner → Executor (calls Tools) → Observer → Reflector → next action
              ↑                                                          │
              └──────────────────────────────────────────────────────────┘
```

## Core Components

### Planner
Decides what to do next. Given a user's goal and conversation history, produces
a plan: a sequence of steps or a next action.

```go
type Planner interface {
    Plan(ctx context.Context, goal string, history []Message) (*Plan, error)
}
```

### Executor
Executes a single step of the plan. May involve calling an LLM, invoking a tool,
or interacting with memory.

```go
type Executor interface {
    Execute(ctx context.Context, step PlanStep, tools []Tool) (*ActionResult, error)
}
```

### Reflection
Evaluates whether the last action moved toward the goal. Decides to continue,
retry, or change strategy.

```go
type Reflector interface {
    Reflect(ctx context.Context, goal string, history []Message, result *ActionResult) (*Reflection, error)
}
```

### Retry
Handles transient failures. Backs off and retries tool calls or LLM requests.

```go
type RetryPolicy struct {
    MaxRetries int
    Backoff    time.Duration
    MaxBackoff time.Duration
}
```

### Tool Calling
The agent has access to a set of `Tool` implementations. The LLM decides which
tool (if any) to call for a given step. See `docs/12-TOOLS.md`.

### Memory Integration
The agent maintains conversation history and can store/retrieve facts via the
`memory.ConversationStore` interface. See Phase 6.

## Agent Loop

```go
// Future API sketch
ai := ihandai.New(
    ihandai.WithLLM("openai", ...),
    ihandai.WithTools(
        tools.NewBrowser(),
        tools.NewCodeExecutor(),
    ),
    ihandai.WithMemory(memory.NewSQLiteStore("agent-memory.db")),
)

goal := "Research the latest Go AI libraries and write a comparison report"
result, err := ai.Run(ctx, goal)
// Agent runs autonomously until goal is reached or max steps exceeded
```

## Agent Patterns (Planned)

| Pattern | Description | Use Case |
|---------|-------------|----------|
| **ReAct** | Reason + Act loop. LLM thinks → acts → observes → repeats | General purpose |
| **Plan & Execute** | Plan first, then execute each step. Split planning from execution | Complex multi-step tasks |
| **Reflection** | After each action, critique and adjust | Tasks requiring quality control |
| **Tool Use** | LLM selects and calls tools from a registry | Data retrieval, external actions |

## Context & Cancellation

All agent operations accept `ctx context.Context`:

- **Timeout**: `ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)` — agent runs for max 5 minutes
- **Cancellation**: user can cancel a runaway agent mid-execution
- **Tracing**: each agent step is a span for OpenTelemetry

## Dependencies

| Dependency | Phase | Status |
|------------|-------|--------|
| `ChatCompleter` (LLM) | Phase 2-3 | Required |
| `Tool` interface | Phase 2 | Required |
| `ConversationStore` (Memory) | Phase 6 | Required |
| `Tool` implementations | Phase 7 | Browser, Code, HTTP, Slack, Email |
| MCP resources | Phase 9 | Optional enhancement |
