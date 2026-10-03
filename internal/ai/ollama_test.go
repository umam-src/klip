package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOllamaCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/api/tags" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider := &Ollama{BaseURL: server.URL}
	if err := provider.Check(context.Background()); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestOllamaListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/tags" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"models":[{"name":"qwen2.5:0.5b"},{"name":"llama3.2:3b"}]}`))
	}))
	defer server.Close()

	provider := &Ollama{BaseURL: server.URL}
	models, err := provider.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels() error = %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("len(models) = %d, want 2", len(models))
	}
	if models[0].ID != "qwen2.5:0.5b" || models[1].ID != "llama3.2:3b" {
		t.Fatalf("models = %#v", models)
	}
}

func TestOllamaCheckRejectsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	provider := &Ollama{BaseURL: server.URL}
	if err := provider.Check(context.Background()); err == nil {
		t.Fatal("Check() error = nil, want provider error")
	}
}
