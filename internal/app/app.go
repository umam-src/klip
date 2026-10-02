package app

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/umam-src/klip/internal/agent"
	"github.com/umam-src/klip/internal/ai"
	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/storage"
)

type App struct {
	config   config.Config
	provider ai.AIProvider
	repo     *storage.Repository
	executor agent.Executor
}

func New(cfg config.Config, provider ai.AIProvider, repos ...*storage.Repository) *App {
	var repo *storage.Repository
	if len(repos) > 0 {
		repo = repos[0]
	}
	executor := agent.Executor{Runner: agent.NewRunner(0, 0), Repo: repo}
	return &App{config: cfg, provider: provider, repo: repo, executor: executor}
}

func (a *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", a.handleHealth)
	mux.HandleFunc("/api/v1/chat", a.handleChat)
	mux.HandleFunc("/api/v1/ruang", a.handleRuang)
	mux.HandleFunc("/api/v1/ruang/", a.handleRuangChild)
	mux.HandleFunc("/api/v1/pekerjaan/", a.handlePekerjaanChild)
	return mux
}

func (a *App) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("ok\n"))
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
