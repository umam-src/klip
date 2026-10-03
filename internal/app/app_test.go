package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/umam-src/klip/internal/agent"
	"github.com/umam-src/klip/internal/ai"
	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

type fakeProvider struct{}

func (fakeProvider) ID() string { return "test" }

func (fakeProvider) Chat(_ context.Context, req ai.ChatRequest) (ai.ChatResponse, error) {
	return ai.ChatResponse{Model: req.Model, Content: "jawaban uji"}, nil
}

type fakeHealthProvider struct {
	err error
}

func (fakeHealthProvider) ID() string { return "health-test" }

func (fakeHealthProvider) Chat(_ context.Context, req ai.ChatRequest) (ai.ChatResponse, error) {
	return ai.ChatResponse{Model: req.Model, Content: "jawaban uji"}, nil
}

func (p fakeHealthProvider) Check(context.Context) error { return p.err }

type fakeModelProvider struct {
	models []ai.Model
	err    error
}

func (fakeModelProvider) ID() string { return "model-test" }

func (fakeModelProvider) Chat(_ context.Context, req ai.ChatRequest) (ai.ChatResponse, error) {
	return ai.ChatResponse{Model: req.Model, Content: "jawaban uji"}, nil
}

func (p fakeModelProvider) ListModels(context.Context) ([]ai.Model, error) {
	return p.models, p.err
}

func BenchmarkAppMemoryIdle(b *testing.B) {
	dataDir := b.TempDir()
	cfg := config.Default(dataDir)
	cfg.AI.Provider = "ollama"
	cfg.AI.Model = "model-uji"

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		app := New(cfg, fakeProvider{})
		runtime.GC()
		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		b.ReportMetric(float64(stats.HeapAlloc), "heap-alloc-bytes")
		runtime.KeepAlive(app)
	}
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

func TestHandlerProviderStatus(t *testing.T) {
	tests := []struct {
		name       string
		provider   ai.AIProvider
		model      string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "reachable",
			provider:   fakeHealthProvider{},
			model:      "model-uji",
			wantStatus: http.StatusOK,
			wantBody:   `{"provider":"ollama","configured":true,"reachable":true}`,
		},
		{
			name:       "unreachable",
			provider:   fakeHealthProvider{err: errors.New("provider down")},
			model:      "model-uji",
			wantStatus: http.StatusServiceUnavailable,
			wantBody:   `{"provider":"ollama","configured":true,"reachable":false}`,
		},
		{
			name:       "provider without health checker",
			provider:   fakeProvider{},
			model:      "model-uji",
			wantStatus: http.StatusOK,
			wantBody:   `{"provider":"ollama","configured":true,"reachable":false}`,
		},
		{
			name:       "provider not configured",
			provider:   fakeHealthProvider{},
			model:      "",
			wantStatus: http.StatusOK,
			wantBody:   `{"provider":"ollama","configured":false,"reachable":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.Default(t.TempDir())
			cfg.AI.Provider = "ollama"
			cfg.AI.Model = tt.model
			handler := New(cfg, tt.provider).Handler()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/provider/status", nil)
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)
			if res.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", res.Code, tt.wantStatus, res.Body.String())
			}
			if got := strings.TrimSpace(res.Body.String()); got != tt.wantBody {
				t.Fatalf("body = %q, want %q", got, tt.wantBody)
			}
		})
	}
}

func TestHandlerProviderModels(t *testing.T) {
	cfg := config.Default(t.TempDir())
	handler := New(cfg, fakeModelProvider{models: []ai.Model{{ID: "model-satu"}, {ID: "model-dua"}}}).Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/provider/models", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusOK)
	}
	if got := strings.TrimSpace(res.Body.String()); got != `{"models":[{"id":"model-satu"},{"id":"model-dua"}]}` {
		t.Fatalf("body = %q", got)
	}
}

func TestHandlerProviderModelsRejectsUnsupportedProvider(t *testing.T) {
	handler := New(config.Default(t.TempDir()), fakeProvider{}).Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/provider/models", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusNotImplemented {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusNotImplemented)
	}
}

func TestHandlerProviderModelsRejectsNonGET(t *testing.T) {
	handler := New(config.Default(t.TempDir()), fakeProvider{}).Handler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/provider/models", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusMethodNotAllowed)
	}
}

func TestHandlerProviderStatusWithoutProvider(t *testing.T) {
	cfg := config.Default(t.TempDir())
	cfg.AI.Provider = "ollama"
	cfg.AI.Model = "model-uji"
	handler := New(cfg, nil).Handler()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/provider/status", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusServiceUnavailable)
	}
	if got := strings.TrimSpace(res.Body.String()); got != `{"provider":"ollama","configured":false,"reachable":false}` {
		t.Fatalf("body = %q", got)
	}
}

func TestHandlerProviderStatusRejectsNonGET(t *testing.T) {
	handler := New(config.Default(t.TempDir()), fakeProvider{}).Handler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/provider/status", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", res.Code, http.StatusMethodNotAllowed)
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
	if !strings.Contains(body, `"model":"model-uji"`) || !strings.Contains(body, `"content":"jawaban uji"`) {
		t.Fatalf("response = %s", body)
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

func TestAppExecutorWiring(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	app := New(config.Default(t.TempDir()), fakeProvider{}, repo)

	if app.executor.Repo != repo {
		t.Fatal("executor repository is not wired to app repository")
	}

	mustCreate := func(name string, fn func() error) {
		t.Helper()
		if err := fn(); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}
	mustCreate("ruang", func() error {
		return repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-exec", Name: "Ruang Executor"})
	})
	mustCreate("agen", func() error {
		return repo.CreateAgen(ctx, domain.Agen{ID: "agen-exec", RuangID: "ruang-exec", Name: "Agen Executor"})
	})
	mustCreate("pekerjaan", func() error {
		return repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-exec", RuangID: "ruang-exec", Title: "Pekerjaan Executor", Status: domain.StatusDraft})
	})
	mustCreate("tugas", func() error {
		return repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-exec", PekerjaanID: "pekerjaan-exec", Title: "Tugas Executor", Status: domain.StatusReady})
	})

	program, args := executionTestCommand()
	result, err := app.executor.Execute(ctx, agent.ExecutionRequest{
		PekerjaanID: "pekerjaan-exec",
		TugasID:     func() *domain.ID { id := domain.ID("tugas-exec"); return &id }(),
		AgenID:      "agen-exec",
		Program:     program,
		Arguments:   args,
