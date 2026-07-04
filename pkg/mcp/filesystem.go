package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FilesystemServer is a simple in-process MCP server that exposes
// a directory as resources. It implements the MCP server protocol
// over an in-memory transport for testing and simple use cases.
//
// For production, use a full MCP server implementation or connect
// to external servers via ConnectStdio.
type FilesystemServer struct {
	root string
}

// NewFilesystemServer creates an in-process filesystem MCP server
// that serves files from the given root directory.
func NewFilesystemServer(root string) *FilesystemServer {
	return &FilesystemServer{root: root}
}

// Handle processes a JSON-RPC request and returns the response result.
func (s *FilesystemServer) Handle(ctx context.Context, method string, params json.RawMessage) (json.RawMessage, error) {
	_ = ctx
	switch method {
	case "initialize":
		return s.handleInitialize(params)
	case "resources/list":
		return s.handleListResources()
	case "resources/read":
		return s.handleReadResource(params)
	default:
		return nil, fmt.Errorf("mcp: unknown method %s", method)
	}
}

func (s *FilesystemServer) handleInitialize(params json.RawMessage) (json.RawMessage, error) {
	result := InitializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities: ServerCapabilities{
			Resources: &ResourcesCapability{},
		},
		ServerInfo: ServerInfo{
			Name:    "ihandai-filesystem",
			Version: "0.1.0",
		},
	}
	return json.Marshal(result)
}

func (s *FilesystemServer) handleListResources() (json.RawMessage, error) {
	var resources []Resource
	err := filepath.Walk(s.root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip files with errors
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasPrefix(info.Name(), ".") {
			return nil
		}
		rel, _ := filepath.Rel(s.root, path)
		resources = append(resources, Resource{
			URI:         "file://" + rel,
			Name:        rel,
			Description: fmt.Sprintf("File: %s (%d bytes)", rel, info.Size()),
			MIMEType:    mimeType(path),
		})
		return nil
	})
	if err != nil {
		return nil, err
	}

	result := ListResourcesResult{Resources: resources}
	return json.Marshal(result)
}

func (s *FilesystemServer) handleReadResource(params json.RawMessage) (json.RawMessage, error) {
	var p ReadResourceParams
	if err := json.Unmarshal(params, &p); err != nil {
		return nil, err
	}

	uri := strings.TrimPrefix(p.URI, "file://")
	path := filepath.Join(s.root, uri)

	// Security: prevent path traversal
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	absRoot, _ := filepath.Abs(s.root)
	if !strings.HasPrefix(absPath, absRoot) {
		return nil, fmt.Errorf("mcp: path traversal attempt: %s", uri)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	result := ReadResourceResult{
		Contents: []TextContent{{
			URI:      p.URI,
			MIMEType: mimeType(absPath),
			Text:     string(data),
		}},
	}
	return json.Marshal(result)
}

func mimeType(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go":
		return "text/x-go"
	case ".md", ".markdown":
		return "text/markdown"
	case ".json":
		return "application/json"
	case ".yaml", ".yml":
		return "text/yaml"
	case ".py":
		return "text/x-python"
	case ".js":
		return "text/javascript"
	case ".html":
		return "text/html"
	case ".css":
		return "text/css"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}

// InMemoryTransport connects a client and an in-process server.
// Useful for testing.
type InMemoryTransport struct {
	server  *FilesystemServer
	pending []json.RawMessage
	mu      sync.Mutex
	cond    *sync.Cond
	closed  bool
}

// NewInMemoryTransport creates a transport that talks to the given server.
func NewInMemoryTransport(server *FilesystemServer) *InMemoryTransport {
	t := &InMemoryTransport{server: server}
	t.cond = sync.NewCond(&t.mu)
	return t
}

func (t *InMemoryTransport) Send(msg any) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Process the request and enqueue the response
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return err
	}

	result, err := t.server.Handle(context.Background(), req.Method, req.Params)
	resp := Response{JSONRPC: jsonrpcVersion, ID: req.ID, Result: result}
	if err != nil {
		resp.Error = &RPCError{Code: -1, Message: err.Error()}
	}

	respData, _ := json.Marshal(resp)
	t.pending = append(t.pending, respData)
	t.cond.Signal()
	return nil
}

func (t *InMemoryTransport) Receive() (json.RawMessage, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for len(t.pending) == 0 && !t.closed {
		t.cond.Wait()
	}
	if t.closed {
		return nil, fmt.Errorf("mcp: transport closed")
	}

	msg := t.pending[0]
	t.pending = t.pending[1:]
	return msg, nil
}

func (t *InMemoryTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.closed = true
	t.cond.Broadcast()
	return nil
}
