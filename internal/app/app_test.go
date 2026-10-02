package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/umam-src/klip/internal/ai"
	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/storage"
)

type fakeProvider struct{}

func (fakeProvider) ID() string { return "test" }

func (fakeProvider) Chat(_ context.Context, req ai.ChatRequest) (ai.ChatResponse, error) {
	return ai.ChatResponse{Model: req.Model, Content: "jawaban uji"}, nil
}

func TestHandlerHealth(t *testing.T) {
	handler := New(config.Default(t.TempDir()), fakeProvider{}).Handler()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if got := res.Body.String(); got != "ok\n" {
		t.Fatalf("body = %q, want %q", got, "ok\n")
	}
}

func TestHandlerChat(t *testing.T) {
	cfg := config.Default(t.TempDir())
	cfg.AI.Model = "model-uji"
	handler := New(cfg, fakeProvider{}).Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", strings.NewReader(`{"prompt":"Halo"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	body := res.Body.String()
	if !strings.Contains(body, `"model":"model-uji"`) {
		t.Fatalf("response missing model: %s", body)
	}
	if !strings.Contains(body, `"content":"jawaban uji"`) {
		t.Fatalf("response missing content: %s", body)
	}
}

func TestHandlerChatRejectsEmptyPrompt(t *testing.T) {
	handler := New(config.Default(t.TempDir()), fakeProvider{}).Handler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", strings.NewReader(`{"prompt":"  "}`))
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusBadRequest)
	}
}

func TestHandlerDomainFlow(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	handler := New(config.Default(t.TempDir()), fakeProvider{}, repo).Handler()

	request := func(method, path, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
		return res
	}

	res := request(http.MethodPost, "/api/v1/ruang", `{"id":"ruang-http","name":"Proyek HTTP"}`)
	if res.Code != http.StatusCreated || !strings.Contains(res.Body.String(), `"id":"ruang-http"`) {
		t.Fatalf("create ruang: status=%d body=%s", res.Code, res.Body.String())
	}

	res = request(http.MethodPost, "/api/v1/ruang/ruang-http/agen", `{"id":"agen-http","name":"Agen HTTP","provider_id":"ollama","model_id":"qwen"}`)
	if res.Code != http.StatusCreated || !strings.Contains(res.Body.String(), `"ruang_id":"ruang-http"`) {
		t.Fatalf("create agen: status=%d body=%s", res.Code, res.Body.String())
	}

	res = request(http.MethodPost, "/api/v1/ruang/ruang-http/pekerjaan", `{"id":"pekerjaan-http","title":"Pekerjaan HTTP"}`)
	if res.Code != http.StatusCreated || !strings.Contains(res.Body.String(), `"status":"draft"`) {
		t.Fatalf("create pekerjaan: status=%d body=%s", res.Code, res.Body.String())
	}

	res = request(http.MethodGet, "/api/v1/ruang/ruang-http/pekerjaan", "")
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"id":"pekerjaan-http"`) {
		t.Fatalf("list pekerjaan: status=%d body=%s", res.Code, res.Body.String())
	}
}
