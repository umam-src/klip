package main

import (
	"context"
	"testing"

	"github.com/umam-src/klip/internal/app"
	"github.com/umam-src/klip/internal/config"
	"github.com/umam-src/klip/internal/storage"
)

func BenchmarkStartupComponents(b *testing.B) {
	for i := 0; i < b.N; i++ {
		cfg := config.Default(b.TempDir())
		db, err := storage.Open(context.Background(), ":memory:")
		if err != nil {
			b.Fatal(err)
		}
		repo := storage.NewRepository(db)
		_ = app.New(cfg, nil, repo).Handler()
		_ = db.Close()
	}
}
