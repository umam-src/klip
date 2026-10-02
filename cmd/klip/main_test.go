package main

import (
	"context"
	"testing"

	"github.com/umam-src/klip/internal/app"
	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/storage"
)

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
