package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkOpenSQLite(b *testing.B) {
	b.ReportAllocs()
	dir := b.TempDir()
	path := filepath.Join(dir, "klip.db")
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		db, err := Open(ctx, path)
		if err != nil {
			b.Fatal(err)
		}
		if err := db.Close(); err != nil {
			b.Fatal(err)
		}
		if err := os.Remove(path); err != nil {
			b.Fatal(err)
		}
	}
}
