package storage

import (
	"context"
	"testing"
)

func TestOpenCreatesSchema(t *testing.T) {
	db, err := Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	for _, table := range []string{"pekerjaan", "event"} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&count); err != nil {
			t.Fatalf("cek tabel %s: %v", table, err)
		}
		if count != 1 {
			t.Fatalf("tabel %s tidak ditemukan", table)
		}
	}
}
