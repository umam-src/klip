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
