package storage

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestBackupAndRestoreDatabase(t *testing.T) {
	ctx := context.Background()
	dbPath := filepath.Join(t.TempDir(), "klip.db")
	backupPath := filepath.Join(t.TempDir(), "backup", "klip.db")
	db, err := Open(ctx, dbPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO ruang(id, name, created_at, updated_at) VALUES ('r1', 'Ruang Uji', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')"); err != nil {
		db.Close()
		t.Fatal(err)
	}

	if err := BackupDatabase(ctx, db, backupPath); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	restoredPath := filepath.Join(t.TempDir(), "restored.db")
	if err := RestoreDatabase(backupPath, restoredPath); err != nil {
		t.Fatal(err)
	}

	restored, err := sql.Open("sqlite", restoredPath)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()

	var name string
	if err := restored.QueryRowContext(ctx, "SELECT name FROM ruang WHERE id = 'r1'").Scan(&name); err != nil {
		t.Fatal(err)
	}
	if name != "Ruang Uji" {
		t.Fatalf("restored name = %q, want %q", name, "Ruang Uji")
	}
}

func TestBackupDatabaseRejectsExistingDestination(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	destination := filepath.Join(t.TempDir(), "backup.db")
	if err := BackupDatabase(ctx, db, destination); err != nil {
		t.Fatal(err)
	}
	if err := BackupDatabase(ctx, db, destination); err == nil {
		t.Fatal("expected existing destination to be rejected")
	}
}
