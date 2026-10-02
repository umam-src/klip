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
	"github.com/umam-src/klip/internal/storage"
	"github.com/umam-src/klip/web"
)

type App struct {
	config      config.Config
	provider    ai.AIProvider
	repo        *storage.Repository
	resultStore *storage.ResultStore
	executor    agent.Executor
}

func New(cfg config.Config, provider ai.AIProvider, repos ...*storage.Repository) *App {
	var repo *storage.Repository
	if len(repos) > 0 {
		repo = repos[0]
	}
	resultStore, _ := storage.NewResultStore(cfg.DataDir)
	executor := agent.Executor{Runner: agent.NewRunner(0, 0), Repo: repo}
	return &App{config: cfg, provider: provider, repo: repo, resultStore: resultStore, executor: executor}
}

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", web.Handler())
	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/api/v1/chat", a.handleChat)
	mux.HandleFunc("/api/v1/provider/status", a.handleProviderStatus)
	mux.HandleFunc("/api/v1/settings", a.handleSettings)
	mux.HandleFunc("/api/v1/ruang", a.handleRuang)
	mux.HandleFunc("/api/v1/ruang/{id}/sasaran", a.handleSasaran)
	mux.HandleFunc("/api/v1/ruang/", a.handleRuangChild)
	mux.HandleFunc("/api/v1/pekerjaan/{id}/komentar", a.handleKomentar)
	mux.HandleFunc("/api/v1/pekerjaan/{id}/approval", a.handlePekerjaanApproval)
	mux.HandleFunc("/api/v1/pekerjaan/", a.handlePekerjaanChild)
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

func (a *App) handleProviderStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
		return
	}
	status := providerStatusResponse{Provider: a.config.AI.Provider}
	if a.provider == nil {
		writeJSON(w, http.StatusServiceUnavailable, status)
		return
	}
	status.Configured = strings.TrimSpace(a.config.AI.Model) != ""
	checker, ok := a.provider.(ai.HealthChecker)
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

type settingsResponse struct {
	DataDir   string `json:"data_dir"`
	Listen    string `json:"listen"`
	Locale    string `json:"locale"`
	Provider  string `json:"provider"`
	BaseURL   string `json:"base_url"`
	Model     string `json:"model"`
	APIKeySet bool   `json:"api_key_set"`
}

func (a *App) handleSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
		return
	}
	writeJSON(w, http.StatusOK, settingsResponse{
		DataDir:   a.config.DataDir,
		Listen:    a.config.Listen,
		Locale:    a.config.Locale,
		Provider:  a.config.AI.Provider,
		BaseURL:   a.config.AI.BaseURL,
		Model:     a.config.AI.Model,
		APIKeySet: strings.TrimSpace(a.config.AI.APIKey) != "",
	})
}

type chatRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model,omitempty"`
}

type chatResponse struct {
	Model   string `json:"model"`
	Content string `json:"content"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func (a *App) handleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
		return
	}
	if a.provider == nil {
		writeError(w, http.StatusServiceUnavailable, "AI belum siap; isi model pada config.json")
		return
	}

	var req chatRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if strings.TrimSpace(req.Prompt) == "" {
		writeError(w, http.StatusBadRequest, "prompt tidak boleh kosong")
		return
	}

	model := a.config.AI.Model
	if strings.TrimSpace(req.Model) != "" {
		model = strings.TrimSpace(req.Model)
	}
	result, err := a.provider.Chat(r.Context(), ai.ChatRequest{
		Model:    model,
		Messages: []ai.Message{{Role: "user", Content: req.Prompt}},
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, "gagal menghubungi penyedia AI")
		return
	}
	writeJSON(w, http.StatusOK, chatResponse{Model: result.Model, Content: result.Content})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
