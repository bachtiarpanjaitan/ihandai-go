package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"sync/atomic"
)

// Transport is the interface for MCP communication channels.
// Implementations include stdio (subprocess) and HTTP/SSE.
type Transport interface {
	// Send writes a JSON-RPC message.
	Send(msg any) error
	// Receive reads a JSON-RPC message.
	Receive() (json.RawMessage, error)
	// Close closes the transport.
	Close() error
}

// Client is an MCP client that connects to an MCP server.
type Client struct {
	transport Transport
	server    ServerInfo
	id        atomic.Int64

	resources map[string]Resource
	tools     map[string]Tool
}

// Connect creates a new MCP client connected to a server via the given transport.
func Connect(transport Transport) (*Client, error) {
	c := &Client{
		transport: transport,
		resources: make(map[string]Resource),
		tools:     make(map[string]Tool),
	}

	// Initialize
	initParams, _ := json.Marshal(InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    Capabilities{},
	})

	resp, err := c.call(context.Background(), "initialize", initParams)
	if err != nil {
		return nil, fmt.Errorf("mcp: initialize: %w", err)
	}

	var initResult InitializeResult
	if err := json.Unmarshal(resp, &initResult); err != nil {
		return nil, fmt.Errorf("mcp: parse initialize response: %w", err)
	}
	c.server = initResult.ServerInfo

	// Discover resources
	if initResult.Capabilities.Resources != nil {
		resources, err := c.ListResources(context.Background())
		if err == nil {
			for _, r := range resources {
				c.resources[r.URI] = r
			}
		}
	}

	// Discover tools
	if initResult.Capabilities.Tools != nil {
		tools, err := c.ListTools(context.Background())
		if err == nil {
			for _, t := range tools {
				c.tools[t.Name] = t
			}
		}
	}

	return c, nil
}

// ConnectStdio connects to an MCP server via stdio (subprocess).
// command is the executable, args are command-line arguments.
func ConnectStdio(command string, args ...string) (*Client, error) {
	cmd := exec.Command(command, args...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("mcp: stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("mcp: stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("mcp: start server: %w", err)
	}

	transport := &stdioTransport{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
	}

	return Connect(transport)
}

// ServerInfo returns information about the connected server.
func (c *Client) ServerInfo() ServerInfo { return c.server }

// ListResources lists all available resources from the server.
func (c *Client) ListResources(ctx context.Context) ([]Resource, error) {
	_ = ctx
	resp, err := c.call(ctx, "resources/list", nil)
	if err != nil {
		return nil, err
	}
	var result ListResourcesResult
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("mcp: parse resources: %w", err)
	}
	return result.Resources, nil
}

// ReadResource reads the content of a resource by URI.
func (c *Client) ReadResource(ctx context.Context, uri string) (*ReadResourceResult, error) {
	_ = ctx
	params, _ := json.Marshal(ReadResourceParams{URI: uri})
	resp, err := c.call(ctx, "resources/read", params)
	if err != nil {
		return nil, err
	}
	var result ReadResourceResult
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("mcp: parse resource: %w", err)
	}
	return &result, nil
}

// ListTools lists all available tools from the server.
func (c *Client) ListTools(ctx context.Context) ([]Tool, error) {
	_ = ctx
	resp, err := c.call(ctx, "tools/list", nil)
	if err != nil {
		return nil, err
	}
	var result ListToolsResult
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("mcp: parse tools: %w", err)
	}
	return result.Tools, nil
}

// CallTool invokes a tool on the server.
func (c *Client) CallTool(ctx context.Context, name string, args json.RawMessage) (*CallToolResult, error) {
	_ = ctx
	params, _ := json.Marshal(CallToolParams{Name: name, Arguments: args})
	resp, err := c.call(ctx, "tools/call", params)
	if err != nil {
		return nil, err
	}
	var result CallToolResult
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("mcp: parse tool result: %w", err)
	}
	return &result, nil
}

// Close closes the connection to the MCP server.
func (c *Client) Close() error {
	return c.transport.Close()
}

// call sends a JSON-RPC request and returns the result.
func (c *Client) call(ctx context.Context, method string, params json.RawMessage) (json.RawMessage, error) {
	id := c.id.Add(1)
	req := Request{
		JSONRPC: jsonrpcVersion,
		ID:      id,
		Method:  method,
		Params:  params,
	}

	if err := c.transport.Send(req); err != nil {
		return nil, fmt.Errorf("mcp: send %s: %w", method, err)
	}

	raw, err := c.transport.Receive()
	if err != nil {
		return nil, fmt.Errorf("mcp: receive %s: %w", method, err)
	}

	var resp Response
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("mcp: parse response: %w", err)
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("mcp: %s error: %s (code %d)", method, resp.Error.Message, resp.Error.Code)
	}

	return resp.Result, nil
}

// stdioTransport implements Transport over a subprocess stdin/stdout.
type stdioTransport struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	mu     sync.Mutex
}

func (t *stdioTransport) Send(msg any) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = t.stdin.Write(data)
	return err
}

func (t *stdioTransport) Receive() (json.RawMessage, error) {
	line, err := t.stdout.ReadBytes('\n')
	if err != nil {
		return nil, err
	}
	return json.RawMessage(line), nil
}

func (t *stdioTransport) Close() error {
	t.stdin.Close()
	return t.cmd.Wait()
}
