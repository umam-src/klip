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

type fakeHealthProvider struct{ err error }

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
func (p fakeModelProvider) ListModels(context.Context) ([]ai.Model, error) { return p.models, p.err }

func TestHandlerHealth(t *testing.T) {
	handler := New(config.Default(t.TempDir()), fakeProvider{}).Handler()
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/health", nil))
	if res.Code != http.StatusOK || res.Body.String() != "ok\n" {
		t.Fatalf("response = %d %q", res.Code, res.Body.String())
	}
}

func TestHandlerProviderStatus(t *testing.T) {
	cfg := config.Default(t.TempDir())
	cfg.AI.Provider = "ollama"
	cfg.AI.Model = "model-uji"
	res := httptest.NewRecorder()
	New(cfg, fakeHealthProvider{}).Handler().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/v1/provider/status", nil))
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"reachable":true`) {
		t.Fatalf("response = %d %s", res.Code, res.Body.String())
	}
	cfg.AI.Model = ""
	res = httptest.NewRecorder()
	New(cfg, fakeHealthProvider{err: errors.New("provider down")}).Handler().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/v1/provider/status", nil))
	if res.Code != http.StatusServiceUnavailable || !strings.Contains(res.Body.String(), `"configured":false`) {
		t.Fatalf("response = %d %s", res.Code, res.Body.String())
	}
}

func TestHandlerProviderModels(t *testing.T) {
	cfg := config.Default(t.TempDir())
	handler := New(cfg, fakeModelProvider{models: []ai.Model{{ID: "model-satu"}, {ID: "model-dua"}}}).Handler()
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/api/v1/provider/models", nil))
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), "model-satu") {
		t.Fatalf("response = %d %s", res.Code, res.Body.String())
	}
}

func TestHandlerChat(t *testing.T) {
	cfg := config.Default(t.TempDir())
	cfg.AI.Model = "model-uji"
	handler := New(cfg, fakeProvider{}).Handler()
	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/chat", strings.NewReader(`{"prompt":"Halo"}`))
	req.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"content":"jawaban uji"`) {
		t.Fatalf("response = %d %s", res.Code, res.Body.String())
	}
}

func TestAppExecutorWiring(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := storage.NewRepository(db)
	app := New(config.Default(t.TempDir()), fakeProvider{}, repo)
	if app.executor.Repo != repo {
		t.Fatal("executor repository tidak terhubung")
	}
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-exec", Name: "Ruang Executor"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-exec", RuangID: "ruang-exec", Name: "Agen Executor"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateGoal(ctx, domain.Goal{ID: "goal-exec", RuangID: "ruang-exec", Title: "Goal Executor", Status: domain.GoalStatusActive}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateProyek(ctx, domain.Proyek{ID: "proyek-exec", RuangID: "ruang-exec", GoalID: "goal-exec", Title: "Proyek Executor", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: "tugas-exec", RuangID: "ruang-exec", ProyekID: "proyek-exec", Title: "Tugas Executor", Status: domain.StatusReady}); err != nil {
		t.Fatal(err)
	}
	program, args := executionTestCommand()
	tugasID := domain.ID("tugas-exec")
	result, err := app.executor.Execute(ctx, agent.ExecutionRequest{ProyekID: "proyek-exec", TugasID: &tugasID, AgenID: "agen-exec", Program: program, Arguments: args})
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Eksekusi.Status != domain.StatusCompleted || strings.ReplaceAll(result.Eksekusi.Stdout, "\r\n", "\n") != "klip-test\n" {
		t.Fatalf("eksekusi = %+v", result.Eksekusi)
	}
	stored, err := repo.GetEksekusi(ctx, result.Eksekusi.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != domain.StatusCompleted || strings.ReplaceAll(stored.Stdout, "\r\n", "\n") != "klip-test\n" {
		t.Fatalf("stored = %+v", stored)
	}
}

func executionTestCommand() (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/C", "echo klip-test"}
	}
	return "printf", []string{"klip-test\\n"}
}

func TestHandlerNativeProyekFlow(t *testing.T) {
	ctx := context.Background()
	db, err := storage.Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	repo := storage.NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-http", Name: "Ruang HTTP"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateGoal(ctx, domain.Goal{ID: "goal-http", RuangID: "ruang-http", Title: "Goal HTTP", Status: domain.GoalStatusActive}); err != nil {
		t.Fatal(err)
	}
	handler := New(config.Default(t.TempDir()), fakeProvider{}, repo).Handler()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ruang/ruang-http/proyek", strings.NewReader(`{"id":"proyek-http","goal_id":"goal-http","title":"Proyek HTTP"}`))
	req.Header.Set("Content-Type", "application/json")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusCreated || !strings.Contains(res.Body.String(), `"id":"proyek-http"`) {
		t.Fatalf("create proyek = %d %s", res.Code, res.Body.String())
	}
}
