package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/umam-src/klip/internal/ai"
	"github.com/umam-src/klip/internal/app"
	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/storage"
)

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

	repo := storage.NewRepository(db)
	provider, err := ai.NewProvider(cfg.AI)
	if err != nil {
		log.Printf("AI belum siap: %v", err)
	}

	server := &http.Server{
		Addr:              cfg.Listen,
		Handler:           app.New(cfg, provider, repo).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}

	shutdownCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-shutdownCtx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("gagal menghentikan server dengan rapi: %v", err)
		}
	}()

	log.Printf("Klip — orkestrator AI lokal")
	log.Printf("data: %s", cfg.DataDir)
	log.Printf("menjalankan server di http://%s", cfg.Listen)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server gagal: %v", err)
	}
}
