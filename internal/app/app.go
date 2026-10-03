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
	mux.HandleFunc("/api/v1/ruang/{id}/goal", a.handleGoal)
	mux.HandleFunc("/api/v1/ruang/{id}/proyek", a.handleProyek)
	mux.HandleFunc("/api/v1/ruang/", a.handleRuangChild)
	mux.HandleFunc("/api/v1/tugas/{id}/agen", a.handleTugasAgen)
	mux.HandleFunc("/api/v1/tugas/{id}/komentar", a.handleTugasKomentar)
	mux.HandleFunc("/api/v1/tugas/{id}/approval", a.handleTugasApproval)
	mux.HandleFunc("/api/v1/tugas/", a.handleTugasKomentar)
	mux.HandleFunc("/api/v1/proyek/", a.handleProyekChild)
	mux.HandleFunc("/api/v1/goal/", a.handleGoalDetail)
	mux.HandleFunc("/api/v1/approval/{id}/{action}", a.handleApprovalDecision)
	mux.HandleFunc("/api/v1/hasil/{id}/file", a.handleHasilFile)
	return mux
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
}

type providerStatusResponse struct { Provider string `json:"provider"`; Configured bool `json:"configured"`; Reachable bool `json:"reachable"` }
type providerConfigRequest struct { Provider string `json:"provider"`; BaseURL string `json:"base_url" }

func (a *App) providerForRequest(w http.ResponseWriter, r *http.Request) ai.AIProvider {
	if r.Method == http.MethodGet { return a.provider }
	if r.Method != http.MethodPost || r.Body == http.NoBody { writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung"); return nil }
	var req providerConfigRequest
	if !decodeJSON(w, r, &req) { return nil }
	req.Provider = strings.TrimSpace(req.Provider); req.BaseURL = strings.TrimSpace(req.BaseURL)
	if req.Provider == "" || req.BaseURL == "" { writeError(w, http.StatusBadRequest, "penyedia dan alamat wajib diisi"); return nil }
	provider, err := ai.NewProvider(config.AIConfig{Provider: req.Provider, BaseURL: req.BaseURL})
	if err != nil { writeError(w, http.StatusBadRequest, "konfigurasi penyedia AI tidak valid"); return nil }
	return provider
}

func (a *App) handleProviderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost { writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung"); return }
	provider := a.providerForRequest(w, r)
	if provider == nil { if r.Method == http.MethodGet { writeJSON(w, http.StatusServiceUnavailable, providerStatusResponse{Provider: a.config.AI.Provider}) }; return }
	status := providerStatusResponse{Provider: provider.ID()}
	if r.Method == http.MethodGet { status.Provider = a.config.AI.Provider; status.Configured = strings.TrimSpace(a.config.AI.Model) != "" }
	checker, ok := provider.(ai.HealthChecker)
	if !ok { writeJSON(w, http.StatusOK, status); return }
	checkCtx, cancel := context.WithTimeout(r.Context(), 5*time.Second); defer cancel()
	if err := checker.Check(checkCtx); err != nil { writeJSON(w, http.StatusServiceUnavailable, status); return }
	status.Reachable = true; writeJSON(w, http.StatusOK, status)
}

type providerModelsResponse struct { Models []ai.Model `json:"models"` }

func (a *App) handleProviderModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost { writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung"); return }
	provider := a.providerForRequest(w, r)
	if provider == nil { if r.Method == http.MethodGet { writeError(w, http.StatusServiceUnavailable, "penyedia AI belum siap") }; return }
	lister, ok := provider.(ai.ModelLister)
	if !ok { writeError(w, http.StatusNotImplemented, "penyedia AI tidak mendukung daftar model"); return }
	listCtx, cancel := context.WithTimeout(r.Context(), 10*time.Second); defer cancel()
	models, err := lister.ListModels(listCtx)
	if err != nil { writeError(w, http.StatusBadGateway, "gagal mengambil daftar model"); return }
	writeJSON(w, http.StatusOK, providerModelsResponse{Models: models})
}

type settingsResponse struct { DataDir string `json:"data_dir"`; Listen string `json:"listen"`; Locale string `json:"locale"`; Provider string `json:"provider"`; BaseURL string `json:"base_url"`; Model string `json:"model"`; APIKeySet bool `json:"api_key_set"` }
type settingsUpdateRequest struct { Provider string `json:"provider"`; BaseURL string `json:"base_url"`; Model string `json:"model"` }

func (a *App) handleSettings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, settingsResponse{DataDir: a.config.DataDir, Listen: a.config.Listen, Locale: a.config.Locale, Provider: a.config.AI.Provider, BaseURL: a.config.AI.BaseURL, Model: a.config.AI.Model, APIKeySet: strings.TrimSpace(a.config.AI.APIKey) != ""})
	case http.MethodPut: a.updateSettings(w, r)
	default: writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func (a *App) updateSettings(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil { writeError(w, http.StatusInternalServerError, "penyimpanan pengaturan belum siap"); return }
	var req settingsUpdateRequest
	if !decodeJSON(w, r, &req) { return }
	req.Provider = strings.TrimSpace(req.Provider); req.BaseURL = strings.TrimSpace(req.BaseURL); req.Model = strings.TrimSpace(req.Model)
	if req.Provider == "" || req.BaseURL == "" || req.Model == "" { writeError(w, http.StatusBadRequest, "penyedia, alamat, dan model wajib diisi"); return }
	updated := a.config; updated.AI.Provider = req.Provider; updated.AI.BaseURL = req.BaseURL; updated.AI.Model = req.Model
	provider, err := ai.NewProvider(updated.AI)
	if err != nil { writeError(w, http.StatusBadRequest, "konfigurasi penyedia AI tidak valid"); return }
	if err := a.repo.SaveAISettings(r.Context(), storage.AISettings{Provider: updated.AI.Provider, BaseURL: updated.AI.BaseURL, Model: updated.AI.Model}); err != nil { writeError(w, http.StatusInternalServerError, "gagal menyimpan pengaturan"); return }
	a.config = updated; a.provider = provider
	writeJSON(w, http.StatusOK, settingsResponse{DataDir: updated.DataDir, Listen: updated.Listen, Locale: updated.Locale, Provider: updated.AI.Provider, BaseURL: updated.AI.BaseURL, Model: updated.AI.Model, APIKeySet: strings.TrimSpace(updated.AI.APIKey) != ""})
}

type chatRequest struct { Prompt string `json:"prompt"`; Model string `json:"model,omitempty"`; Stream bool `json:"stream,omitempty"` }
type chatResponse struct { Model string `json:"model"`; Content string `json:"content"` }
type errorResponse struct { Error string `json:"error"` }

func (a *App) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost { writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung"); return }
	if a.provider == nil { writeError(w, http.StatusServiceUnavailable, "AI belum siap; isi model pada config.json"); return }
	var req chatRequest
	if !decodeJSON(w, r, &req) { return }
	if strings.TrimSpace(req.Prompt) == "" { writeError(w, http.StatusBadRequest, "prompt tidak boleh kosong"); return }
	model := a.config.AI.Model; if strings.TrimSpace(req.Model) != "" { model = strings.TrimSpace(req.Model) }
	chatReq := ai.ChatRequest{Model: model, Messages: []ai.Message{{Role: "user", Content: req.Prompt}}}
	if req.Stream { a.handleChatStream(w, r, chatReq); return }
	result, err := a.provider.Chat(r.Context(), chatReq)
	if err != nil { writeError(w, http.StatusBadGateway, "gagal menghubungi penyedia AI"); return }
	writeJSON(w, http.StatusOK, chatResponse{Model: result.Model, Content: result.Content})
}

func (a *App) handleChatStream(w http.ResponseWriter, r *http.Request, req ai.ChatRequest) {
	streamer, ok := a.provider.(ai.Streamer); if !ok { writeError(w, http.StatusNotImplemented, "penyedia AI tidak mendukung streaming"); return }
	flusher, ok := w.(http.Flusher); if !ok { writeError(w, http.StatusInternalServerError, "server tidak mendukung streaming"); return }
	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8"); w.Header().Set("Cache-Control", "no-cache"); w.Header().Set("X-Accel-Buffering", "no"); w.WriteHeader(http.StatusOK)
	emit := func(chunk ai.ChatResponse) error {
		payload, err := json.Marshal(chatResponse{Model: chunk.Model, Content: chunk.Content}); if err != nil { return err }
		if _, err := w.Write([]byte("data: ")); err != nil { return err }; if _, err := w.Write(payload); err != nil { return err }; if _, err := w.Write([]byte("\n\n")); err != nil { return err }; flusher.Flush(); return nil
	}
	if err := streamer.Stream(r.Context(), req, emit); err != nil { return }
	_, _ = w.Write([]byte("data: [DONE]\n\n")); flusher.Flush()
}

func writeJSON(w http.ResponseWriter, status int, value any) { w.Header().Set("Content-Type", "application/json; charset=utf-8"); w.WriteHeader(status); _ = json.NewEncoder(w).Encode(value) }
func writeError(w http.ResponseWriter, status int, message string) { writeJSON(w, status, errorResponse{Error: message}) }
