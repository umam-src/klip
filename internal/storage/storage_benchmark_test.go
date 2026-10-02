package storage

import (
	"context"
	"path/filepath"
	"testing"
)

func BenchmarkOpenSQLite(b *testing.B) {
	for i := 0; i < b.N; i++ {
		path := filepath.Join(b.TempDir(), "klip.db")
		db, err := Open(context.Background(), path)
		if err != nil {
			b.Fatal(err)
		}
		if err := db.Close(); err != nil {
			b.Fatal(err)
		}
	}
}
