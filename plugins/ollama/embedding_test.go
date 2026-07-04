package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestEmbedder_Integration(t *testing.T) {
	url := os.Getenv("OLLAMA_URL")
	if url == "" {
		t.Skip("OLLAMA_URL not set, skipping integration test")
	}

	e := NewEmbedder(WithBaseURL(url), WithModel("nomic-embed-text"))
	vec, err := e.Embed(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	if len(vec) == 0 {
		t.Error("expected non-empty embedding vector")
	}
	t.Logf("Vector dimensions: %d", len(vec))
}

func TestEmbedder_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("unexpected method: %s", r.Method)
		}

		var req embedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		if req.Model != "nomic-embed-text" {
			t.Errorf("Model: got %q, want %q", req.Model, "nomic-embed-text")
		}

		resp := embedResponse{
			Embedding: []float64{0.1, 0.2, 0.3, 0.4, 0.5},
		}
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatalf("failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	e := NewEmbedder(WithBaseURL(server.URL), WithModel("nomic-embed-text"))
	vec, err := e.Embed(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("Embed failed: %v", err)
	}
	if len(vec) != 5 {
		t.Errorf("got %d dimensions, want 5", len(vec))
	}
}

func TestEmbedder_EmbedBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := embedResponse{
			Embedding: []float64{0.1, 0.2, 0.3},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	e := NewEmbedder(WithBaseURL(server.URL), WithModel("nomic-embed-text"))
	vecs, err := e.EmbedBatch(context.Background(), []string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("EmbedBatch failed: %v", err)
	}
	if len(vecs) != 3 {
		t.Errorf("got %d vectors, want 3", len(vecs))
	}
	for i, vec := range vecs {
		if len(vec) != 3 {
			t.Errorf("vector %d: got %d dimensions, want 3", i, len(vec))
		}
	}
}

func TestEmbedder_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	e := NewEmbedder(WithBaseURL(server.URL), WithModel("nomic-embed-text"))
	_, err := e.Embed(context.Background(), "test")
	if err == nil {
		t.Error("expected error for 503 response")
	}
}
