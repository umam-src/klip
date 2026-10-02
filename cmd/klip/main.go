package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
		Addr:    cfg.Listen,
		Handler: app.New(cfg, provider, repo).Handler(),
	}

	log.Printf("Klip — orkestrator AI lokal")
	log.Printf("data: %s", cfg.DataDir)
	log.Printf("menjalankan server di http://%s", cfg.Listen)
	log.Fatal(server.ListenAndServe())
}
