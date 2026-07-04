// Package mcp implements a Model Context Protocol (MCP) client.
//
// MCP (https://modelcontextprotocol.io) is a standard protocol for
// providing resources and prompts to LLMs through a client-server architecture.
// This package implements the client side of the protocol.
package mcp

import "encoding/json"

// JSON-RPC message types used by MCP.
const (
	jsonrpcVersion = "2.0"
)

// Request is a JSON-RPC request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError is a JSON-RPC error.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Notification is a JSON-RPC notification (no ID).
type Notification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// InitializeParams is sent by the client to initialize the connection.
type InitializeParams struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
}

// Capabilities describes what the client supports.
type Capabilities struct{}

// InitializeResult is the server's response to initialize.
type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      ServerInfo         `json:"serverInfo"`
}

// ServerCapabilities describes what the server supports.
type ServerCapabilities struct {
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Tools     *ToolsCapability     `json:"tools,omitempty"`
}

// ResourcesCapability indicates resource support.
type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ToolsCapability indicates tool support.
type ToolsCapability struct{}

// ServerInfo contains server metadata.
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Resource represents an MCP resource.
type Resource struct {
	URI         string `json:"uri"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	MIMEType    string `json:"mimeType,omitempty"`
}

// ListResourcesResult is the response to resources/list.
type ListResourcesResult struct {
	Resources []Resource `json:"resources"`
}

// ReadResourceParams is the request to read a resource.
type ReadResourceParams struct {
	URI string `json:"uri"`
}

// ReadResourceResult is the response to resources/read.
type ReadResourceResult struct {
	Contents []TextContent `json:"contents"`
}

// TextContent represents a text resource content.
type TextContent struct {
	URI      string `json:"uri"`
	MIMEType string `json:"mimeType,omitempty"`
	Text     string `json:"text"`
}

// Tool represents an MCP tool.
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ListToolsResult is the response to tools/list.
type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

// CallToolParams is the request to call a tool.
type CallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// CallToolResult is the response to tools/call.
type CallToolResult struct {
	Content []TextContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}
