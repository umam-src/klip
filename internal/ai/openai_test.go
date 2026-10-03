package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAICompatibleChat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"local-model","choices":[{"message":{"role":"assistant","content":"Halo dari AI lokal"}}]}`))
	}))
	defer server.Close()

	provider := &OpenAICompatible{BaseURL: server.URL, APIKey: "test-key"}
	got, err := provider.Chat(context.Background(), ChatRequest{
		Model:    "local-model",
		Messages: []Message{{Role: "user", Content: "Halo"}},
	})
	if err != nil {
		t.Fatalf("Chat() error = %v", err)
	}
	if got.Content != "Halo dari AI lokal" {
		t.Fatalf("content = %q", got.Content)
	}
}

func TestOpenAICompatibleStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: {\"model\":\"local-model\",\"choices\":[{\"delta\":{\"content\":\"Halo \"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"model\":\"local-model\",\"choices\":[{\"delta\":{\"content\":\"dunia\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	provider := &OpenAICompatible{BaseURL: server.URL}
	var got strings.Builder
	err := provider.Stream(context.Background(), ChatRequest{Model: "local-model"}, func(chunk ChatResponse) error {
		got.WriteString(chunk.Content)
		return nil
	})
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}
	if got.String() != "Halo dunia" {
		t.Fatalf("content = %q", got.String())
	}
}

func TestOpenAICompatibleListModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/v1/models" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization = %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"qwen2.5-0.5b-instruct-q4_k_m"},{"id":"model-kedua"}]}`))
	}))
	defer server.Close()

	provider := &OpenAICompatible{BaseURL: server.URL, APIKey: "test-key"}
	models, err := provider.ListModels(context.Background())
	if err != nil {
		t.Fatalf("ListModels() error = %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("len(models) = %d, want 2", len(models))
	}
	if models[0].ID != "qwen2.5-0.5b-instruct-q4_k_m" || models[1].ID != "model-kedua" {
		t.Fatalf("models = %#v", models)
	}
}

func TestOpenAICompatibleCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s", r.Method)
		}
		if r.URL.Path != "/v1/models" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("authorization = %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	provider := &OpenAICompatible{BaseURL: server.URL, APIKey: "test-key"}
	if err := provider.Check(context.Background()); err != nil {
		t.Fatalf("Check() error = %v", err)
	}
}

func TestOpenAICompatibleCheckRejectsHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	provider := &OpenAICompatible{BaseURL: server.URL}
	if err := provider.Check(context.Background()); err == nil {
		t.Fatal("Check() error = nil, want provider error")
	}
}
