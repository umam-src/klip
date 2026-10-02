package main

import (
	"context"
	"flag"
	"fmt"
	"io"
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

var version = "dev"

func runCLI(args []string, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		return serve(args, stderr)
	}

	switch args[0] {
	case "help", "--help", "-h":
		writeUsage(stdout)
		return nil
	case "version", "--version":
		_, err := fmt.Fprintf(stdout, "klip %s\n", version)
		return err
	case "serve":
		return serve(args[1:], stderr)
	default:
		return fmt.Errorf("perintah tidak dikenal: %q; gunakan 'klip help'", args[0])
	}
}

func serve(args []string, stderr io.Writer) error {
	flags := flag.NewFlagSet("klip serve", flag.ContinueOnError)
	flags.SetOutput(stderr)
	dataDirFlag := flags.String("data-dir", "", "direktori data Klip")
	listenFlag := flags.String("listen", "", "alamat server HTTP")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("argumen tidak dikenal: %q", flags.Arg(0))
	}

	dataDir, err := defaultDataDir(*dataDirFlag)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return fmt.Errorf("gagal membuat direktori data: %w", err)
	}

	cfg, err := config.Load(filepath.Join(dataDir, "config.json"), config.Default(dataDir))
	if err != nil {
		return fmt.Errorf("gagal membaca konfigurasi: %w", err)
	}
	if *dataDirFlag != "" {
		cfg.DataDir = dataDir
	}
	if *listenFlag != "" {
		cfg.Listen = *listenFlag
	}

	ctx := context.Background()
	db, err := storage.Open(ctx, filepath.Join(cfg.DataDir, "klip.db"))
	if err != nil {
		return fmt.Errorf("gagal membuka database: %w", err)
	}
	defer db.Close()

	repo := storage.NewRepository(db)
	provider, err := ai.NewProvider(cfg.AI)
	if err != nil {
		log.Printf("AI belum siap: %v", err)
	}

	application := app.New(cfg, provider, repo)
	server := &http.Server{
		Addr:              cfg.Listen,
		Handler:           application.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      2 * time.Minute,
		IdleTimeout:       60 * time.Second,
	}

	shutdownCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	application.Start(shutdownCtx)

	go func() {
		<-shutdownCtx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		application.Stop()
		if err := server.Shutdown(shutdown); err != nil {
			log.Printf("gagal menghentikan server dengan rapi: %v", err)
		}
	}()

	log.Printf("Klip — orkestrator AI lokal")
	log.Printf("data: %s", cfg.DataDir)
	log.Printf("menjalankan server di http://%s", cfg.Listen)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server gagal: %w", err)
	}
	return nil
}

func defaultDataDir(explicit string) (string, error) {
	if explicit != "" {
		return filepath.Clean(explicit), nil
	}
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("gagal menentukan direktori data: %w", err)
	}
	return filepath.Join(baseDir, "Klip"), nil
}

func writeUsage(w io.Writer) {
	_, _ = fmt.Fprint(w, `Klip — orkestrator AI lokal

Penggunaan:
  klip [perintah] [opsi]

Perintah:
  serve      jalankan server lokal (default)
  version    tampilkan versi Klip
  help       tampilkan bantuan ini

Opsi serve:
  --data-dir PATH   gunakan direktori data tertentu
  --listen ADDR     gunakan alamat server tertentu

Contoh:
  klip
  klip serve --data-dir ./data
  klip serve --listen 127.0.0.1:8788
  klip version
`)
}
