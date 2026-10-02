package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/umam-src/klip/internal/ai"
	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/storage"
)

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

func main() {
	ctx := context.Background()

	baseDir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("gagal menentukan direktori data: %v", err)
	}
	dataDir := filepath.Join(baseDir, "Klip")
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		log.Fatalf("gagal membuat direktori data: %v", err)
	}

	cfg, err := config.Load(filepath.Join(dataDir, "config.json"), config.Default(dataDir))
	if err != nil {
		log.Fatalf("gagal membaca konfigurasi: %v", err)
	}

	db, err := storage.Open(ctx, filepath.Join(cfg.DataDir, "klip.db"))
	if err != nil {
		log.Fatalf("gagal membuka database: %v", err)
	}
	defer db.Close()

	provider, err := ai.NewProvider(cfg.AI)
	if err != nil {
		log.Printf("AI belum siap: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})
	mux.HandleFunc("/api/v1/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
			return
		}
		if provider == nil {
			writeError(w, http.StatusServiceUnavailable, "AI belum siap; isi model pada config.json")
			return
		}

		var req chatRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "permintaan tidak valid")
			return
		}
		if strings.TrimSpace(req.Prompt) == "" {
			writeError(w, http.StatusBadRequest, "prompt tidak boleh kosong")
			return
		}

		model := cfg.AI.Model
		if strings.TrimSpace(req.Model) != "" {
			model = strings.TrimSpace(req.Model)
		}
		result, err := provider.Chat(r.Context(), ai.ChatRequest{
			Model: model,
			Messages: []ai.Message{{
				Role:    "user",
				Content: req.Prompt,
			}},
		})
		if err != nil {
			writeError(w, http.StatusBadGateway, "gagal menghubungi penyedia AI")
			log.Printf("chat gagal: %v", err)
			return
		}

		writeJSON(w, http.StatusOK, chatResponse{Model: result.Model, Content: result.Content})
	})

	fmt.Println("Klip — orkestrator AI lokal")
	log.Printf("data: %s", cfg.DataDir)
	log.Printf("menjalankan server di http://%s", cfg.Listen)
	log.Fatal(http.ListenAndServe(cfg.Listen, mux))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
