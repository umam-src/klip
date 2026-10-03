package storage

import (
	"context"
	"database/sql"
	"testing"
)

// newTestDB membuka database SQLite in-memory yang ditutup otomatis saat test selesai.
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// mustExec menjalankan pernyataan SQL dan menghentikan test jika gagal.
func mustExec(t *testing.T, db *sql.DB, query string, args ...any) {
	t.Helper()
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("exec %q error = %v", query, err)
	}
}
