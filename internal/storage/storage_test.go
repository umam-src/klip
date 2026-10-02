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

    var count int
    if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'pekerjaan'").Scan(&count); err != nil {
        t.Fatalf("cek tabel pekerjaan: %v", err)
    }
    if count != 1 {
        t.Fatalf("tabel pekerjaan tidak ditemukan")
    }
}
