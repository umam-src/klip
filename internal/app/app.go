package app

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/umam-src/klip/internal/agent"
	"github.com/umam-src/klip/internal/ai"
	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/scheduler"
	"github.com/umam-src/klip/internal/storage"
	"github.com/umam-src/klip/web"
)

type App struct {
	config      config.Config
	provider    ai.AIProvider
	repo        *storage.Repository
	resultStore *storage.ResultStore
	executor    agent.Executor
	scheduler   *scheduler.Scheduler
}

func New(cfg config.Config, provider ai.AIProvider, repos ...*storage.Repository) *App {
	var repo *storage.Repository
	if len(repos) > 0 {
		repo = repos[0]
	}
	resultStore, _ := storage.NewResultStore(cfg.DataDir)
	executor := agent.Executor{Runner: agent.NewRunner(0, 0), Repo: repo}
	app := &App{config: cfg, provider: provider, repo: repo, resultStore: resultStore, executor: executor}
	if repo != nil {
		app.scheduler, _ = scheduler.New(repo, executor, scheduler.Config{})
	}
	return app
}

func (a *App) Start(ctx context.Context) {
	if a.scheduler != nil {
		a.scheduler.Start(ctx)
	}
}

func (a *App) Stop() {
	if a.scheduler != nil {
		a.scheduler.Stop()
	}
}

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", web.Handler())
	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/api/v1/chat", a.handleChat)
	mux.HandleFunc("/api/v1/provider/status", a.handleProviderStatus)
	mux.HandleFunc("/api/v1/provider/models", a.handleProviderModels)
	mux.HandleFunc("/api/v1/settings", a.handleSettings)
	mux.HandleFunc("/api/v1/scheduler", a.handleScheduler)
	mux.HandleFunc("/api/v1/scheduler/", a.handleSchedulerChild)
	mux.HandleFunc("/api/v1/ruang", a.handleRuang)
	mux.HandleFunc("/api/v1/ruang/{id}/sasaran", a.handleSasaran)
	mux.HandleFunc("/api/v1/ruang/", a.handleRuangChild)
	mux.HandleFunc("/api/v1/pekerjaan/{id}/komentar", a.handleKomentar)
	mux.HandleFunc("/api/v1/pekerjaan/{id}/approval", a.handlePekerjaanApproval)
	mux.HandleFunc("/api/v1/pekerjaan/", a.handlePekerjaanChild)
	mux.HandleFunc("/api/v1/tugas/{id}/agen", a.handleTugasAgen)
	mux.HandleFunc("/api/v1/tugas/{id}/komentar", a.handleTugasKomentar)
	mux.HandleFunc("/api/v1/tugas/{id}/approval", a.handleTugasApproval)
	mux.HandleFunc("/api/v1/tugas/", a.handleTugasKomentar)
	mux.HandleFunc("/api/v1/approval/{id}/{action}", a.handleApprovalDecision)
	mux.HandleFunc("/api/v1/hasil/{id}/file", a.handleHasilFile)
	return mux
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

type providerStatusResponse struct {
	Provider   string `json:"provider"`
	Configured bool   `json:"configured"`
	Reachable  bool   `json:"reachable"`
}

type providerConfigRequest struct {
	Provider string `json:"provider"`
	BaseURL  string `json:"base_url"`
}

func (a *App) providerForRequest(w http.ResponseWriter, r *http.Request) ai.AIProvider {
	if r.Method == http.MethodGet {
		return a.provider
	}
	if r.Method != http.MethodPost || r.Body == http.NoBody {
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
		return nil
	}
	var req providerConfigRequest
	if !decodeJSON(w, r, &req) {
		return nil
	}
	req.Provider = strings.TrimSpace(req.Provider)
	req.BaseURL = strings.TrimSpace(req.BaseURL)
	if req.Provider == "" || req.BaseURL == "" {
		writeError(w, http.StatusBadRequest, "penyedia dan alamat wajib diisi")
		return nil
	}
	provider, err := ai.NewProvider(config.AIConfig{Provider: req.Provider, BaseURL: req.BaseURL})
	if err != nil {
		writeError(w, http.StatusBadRequest, "konfigurasi penyedia AI tidak valid")
		return nil
	}
	return provider
}

func (a *App) handleProviderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
		return
	}
	provider := a.providerForRequest(w, r)
	if provider == nil {
		if r.Method == http.MethodGet {
			writeJSON(w, http.StatusServiceUnavailable, providerStatusResponse{Provider: a.config.AI.Provider})
		}
		return
	}
	status := providerStatusResponse{Provider: provider.ID()}
	if r.Method == http.MethodGet {
		status.Provider = a.config.AI.Provider
		status.Configured = strings.TrimSpace(a.config.AI.Model) != ""
	}
	checker, ok := provider.(ai.HealthChecker)
	if !ok {
		writeJSON(w, http.StatusOK, status)
		return
	}
	checkCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	if err := checker.Check(checkCtx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, status)
		return
	}
	status.Reachable = true
	writeJSON(w, http.StatusOK, status)
}

type providerModelsResponse struct {
	Models []ai.Model `json:"models"`
}

func (a *App) handleProviderModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
		return
	}
	provider := a.providerForRequest(w, r)
	if provider == nil {
		if r.Method == http.MethodGet {
			writeError(w, http.StatusServiceUnavailable, "penyedia AI belum siap")
		}
		return
	}
	lister, ok := provider.(ai.ModelLister)
	if !ok {
		writeError(w, http.StatusNotImplemented, "penyedia AI tidak mendukung daftar model")
		return
	}
	listCtx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	models, err := lister.ListModels(listCtx)
	if err != nil {
		writeError(w, http.StatusBadGateway, "gagal mengambil daftar model")
		return
	}
	writeJSON(w, http.StatusOK, providerModelsResponse{Models: models})
}
