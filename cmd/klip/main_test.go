package main

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/umam-src/klip/internal/app"
	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/storage"
)

func TestRunCLIVersion(t *testing.T) {
	var out, errOut bytes.Buffer
	old := version
	version = "0.1.0"
	t.Cleanup(func() { version = old })

	if err := runCLI([]string{"version"}, &out, &errOut); err != nil {
		t.Fatalf("runCLI() error = %v", err)
	}
	if got := out.String(); got != "klip 0.1.0\n" {
		t.Fatalf("output = %q", got)
	}
	if errOut.Len() != 0 {
		t.Fatalf("stderr = %q", errOut.String())
	}
}

func TestRunCLIHelp(t *testing.T) {
	var out, errOut bytes.Buffer
	if err := runCLI([]string{"help"}, &out, &errOut); err != nil {
		t.Fatalf("runCLI() error = %v", err)
	}
	if !strings.Contains(out.String(), "klip serve") {
		t.Fatalf("help output = %q", out.String())
	}
	if errOut.Len() != 0 {
		t.Fatalf("stderr = %q", errOut.String())
	}
}

func TestRunCLIRejectsUnknownCommand(t *testing.T) {
	var out, errOut bytes.Buffer
	err := runCLI([]string{"tidak-ada"}, &out, &errOut)
	if err == nil {
		t.Fatal("runCLI() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "perintah tidak dikenal") {
		t.Fatalf("error = %v", err)
	}
}

func TestDefaultDataDir(t *testing.T) {
	got, err := defaultDataDir("./data")
	if err != nil {
		t.Fatalf("defaultDataDir() error = %v", err)
	}
	if got != "data" {
		t.Fatalf("data dir = %q, want %q", got, "data")
	}
}

func BenchmarkStartupComponents(b *testing.B) {
	b.ReportAllocs()
	dataDir := b.TempDir()
	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		cfg := config.Default(dataDir)
		db, err := storage.Open(ctx, ":memory:")
		if err != nil {
			b.Fatal(err)
		}
		repo := storage.NewRepository(db)
		_ = app.New(cfg, nil, repo).Handler()
		if err := db.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
