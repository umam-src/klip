package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/umam-src/klip/internal/ai"
	"github.com/umam-src/klip/internal/config"
)

type fakeStreamProvider struct{}

func (fakeStreamProvider) ID() string { return "stream-test" }

func (fakeStreamProvider) Chat(context.Context, ai.ChatRequest) (ai.ChatResponse, error) {
	return ai.ChatResponse{}, nil
}

func (fakeStreamProvider) Stream(_ context.Context, req ai.ChatRequest, emit func(ai.ChatResponse) error) error {
	if err := emit(ai.ChatResponse{Model: req.Model, Content: "Halo "}); err != nil {
		return err
	}
	return emit(ai.ChatResponse{Model: req.Model, Content: "dunia"})
}

func TestHandlerChatStream(t *testing.T) {
	cfg := config.Default(t.TempDir())
	cfg.AI.Model = "model-uji"
	handler := New(cfg, fakeStreamProvider{}).Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", strings.NewReader(`{"prompt":"Halo","stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", res.Code, http.StatusOK, res.Body.String())
	}
	if got := res.Header().Get("Content-Type"); got != "text/event-stream; charset=utf-8" {
		t.Fatalf("content type = %q", got)
	}
	body := res.Body.String()
	if !strings.Contains(body, `data: {"model":"model-uji","content":"Halo "}`) ||
		!strings.Contains(body, `data: {"model":"model-uji","content":"dunia"}`) ||
		!strings.HasSuffix(body, "data: [DONE]\n\n") {
		t.Fatalf("stream body = %q", body)
	}
}

func TestHandlerChatStreamUnsupported(t *testing.T) {
	cfg := config.Default(t.TempDir())
	cfg.AI.Model = "model-uji"
	handler := New(cfg, fakeProvider{}).Handler()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", strings.NewReader(`{"prompt":"Halo","stream":true}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	if res.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d; body = %s", res.Code, http.StatusNotImplemented, res.Body.String())
	}
}
