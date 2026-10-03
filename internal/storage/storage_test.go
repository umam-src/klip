package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenCreatesSchemaV1(t *testing.T) {
	db, err := Open(context.Background(), ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	for _, table := range []string{
		"ruang",
		"goal",
		"proyek",
		"tugas",
		"penugasan",
		"agen",
		"eksekusi",
		"peristiwa",
		"hasil",
		"komentar",
		"persetujuan",
		"jadwal",
		"jadwal_eksekusi",
		"pengaturan_ai",
	} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&count); err != nil {
			t.Fatalf("cek tabel %s: %v", table, err)
		}
		if count != 1 {
			t.Fatalf("tabel %s tidak ditemukan", table)
		}
	}

	for _, table := range []string{"pekerjaan", "sasaran", "run", "sesi", "event", "tugas_agen"} {
		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", table).Scan(&count); err != nil {
			t.Fatalf("cek tabel legacy %s: %v", table, err)
		}
		if count != 0 {
			t.Fatalf("tabel legacy %s masih dibuat", table)
		}
	}

	var version int
	if err := db.QueryRow("SELECT MAX(version) FROM schema_migrations").Scan(&version); err != nil {
		t.Fatalf("baca versi skema: %v", err)
	}
	if version != schemaVersion {
		t.Fatalf("versi skema = %d, want %d", version, schemaVersion)
	}
}

func TestOpenRejectsLegacyDatabase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "klip.db")

	db, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("buat database v1: %v", err)
	}
	if _, err := db.Exec("DROP TABLE ruang; CREATE TABLE pekerjaan (id TEXT PRIMARY KEY)"); err != nil {
		db.Close()
		t.Fatalf("buat fixture legacy: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("tutup database: %v", err)
	}

	if _, err := Open(context.Background(), path); err == nil || !strings.Contains(err.Error(), "tidak didukung") {
		t.Fatalf("Open() error = %v, want database legacy ditolak", err)
	}
}

func TestOpenDoesNotRewriteDatabaseWithUnsupportedVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "klip.db")

	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}

	db, err := Open(context.Background(), path)
	if err != nil {
		t.Fatalf("buat database: %v", err)
	}
	if _, err := db.Exec("UPDATE schema_migrations SET version = 99"); err != nil {
		db.Close()
		t.Fatalf("ubah versi skema: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	if _, err := Open(context.Background(), path); err == nil || !strings.Contains(err.Error(), "lebih baru") {
		t.Fatalf("Open() error = %v, want versi lebih baru ditolak", err)
	}
}
