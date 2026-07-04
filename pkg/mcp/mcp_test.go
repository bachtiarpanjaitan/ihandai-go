package mcp

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestFilesystemServer_ListResources(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("hello"), 0644)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("# Title"), 0644)

	srv := NewFilesystemServer(dir)
	transport := NewInMemoryTransport(srv)

	client, err := Connect(transport)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	resources, err := client.ListResources(context.Background())
	if err != nil {
		t.Fatalf("ListResources failed: %v", err)
	}
	if len(resources) != 2 {
		t.Errorf("got %d resources, want 2", len(resources))
	}

	names := make(map[string]bool)
	for _, r := range resources {
		names[r.Name] = true
	}
	if !names["test.txt"] || !names["README.md"] {
		t.Errorf("missing expected resources in %v", names)
	}
}

func TestFilesystemServer_ReadResource(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "hello.txt"), []byte("hello world"), 0644)

	srv := NewFilesystemServer(dir)
	transport := NewInMemoryTransport(srv)

	client, err := Connect(transport)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	result, err := client.ReadResource(context.Background(), "file://hello.txt")
	if err != nil {
		t.Fatalf("ReadResource failed: %v", err)
	}
	if len(result.Contents) != 1 {
		t.Fatalf("got %d contents, want 1", len(result.Contents))
	}
	if result.Contents[0].Text != "hello world" {
		t.Errorf("got %q, want %q", result.Contents[0].Text, "hello world")
	}
	if result.Contents[0].MIMEType != "text/plain" {
		t.Errorf("got MIME %q, want text/plain", result.Contents[0].MIMEType)
	}
}

func TestFilesystemServer_PathTraversal(t *testing.T) {
	dir := t.TempDir()

	srv := NewFilesystemServer(dir)
	transport := NewInMemoryTransport(srv)

	client, err := Connect(transport)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	_, err = client.ReadResource(context.Background(), "file://../../../etc/passwd")
	if err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestFilesystemServer_ServerInfo(t *testing.T) {
	srv := NewFilesystemServer("/tmp")
	transport := NewInMemoryTransport(srv)

	client, err := Connect(transport)
	if err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Close()

	if client.ServerInfo().Name != "ihandai-filesystem" {
		t.Errorf("server name: got %q, want ihandai-filesystem", client.ServerInfo().Name)
	}
}

func TestProtocol_Types(t *testing.T) {
	req := Request{JSONRPC: "2.0", ID: 1, Method: "initialize"}
	if req.JSONRPC != "2.0" {
		t.Error("JSONRPC version mismatch")
	}

	resp := Response{JSONRPC: "2.0", ID: 1, Error: &RPCError{Code: -1, Message: "err"}}
	if resp.Error.Code != -1 {
		t.Error("error code mismatch")
	}
}

func TestFilesystemServer_MIMETypes(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"test.go", "text/x-go"},
		{"test.md", "text/markdown"},
		{"test.json", "application/json"},
		{"test.py", "text/x-python"},
		{"test.txt", "text/plain"},
		{"test.unknown", "application/octet-stream"},
	}
	for _, tt := range tests {
		got := mimeType(tt.path)
		if got != tt.want {
			t.Errorf("mimeType(%q): got %q, want %q", tt.path, got, tt.want)
		}
	}
}
