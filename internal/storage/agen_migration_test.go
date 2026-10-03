package storage

import (
	"context"
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func TestMigrateAgenRoleStatusFromV9(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	defer db.Close()

	for _, statement := range []string{
		`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)`,
		`INSERT INTO schema_migrations(version) VALUES (9)`,
		`CREATE TABLE ruang (id TEXT PRIMARY KEY, name TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE TABLE agen (id TEXT PRIMARY KEY, ruang_id TEXT NOT NULL, parent_id TEXT, name TEXT NOT NULL, description TEXT NOT NULL DEFAULT '', provider_id TEXT NOT NULL DEFAULT '', model_id TEXT NOT NULL DEFAULT '', created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`INSERT INTO agen (id, ruang_id, name, created_at, updated_at) VALUES ('agen-legacy', 'ruang-1', 'Legacy', '2026-10-02T05:00:00Z', '2026-10-02T05:00:00Z')`,
	} {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatalf("setup statement %q: %v", statement, err)
		}
	}

	if err := migrate(ctx, db); err != nil {
		t.Fatalf("migrate() error = %v", err)
	}

	var role, status string
	if err := db.QueryRowContext(ctx, `SELECT role, status FROM agen WHERE id = 'agen-legacy'`).Scan(&role, &status); err != nil {
		t.Fatalf("read migrated agent: %v", err)
	}
	if role != "Agen" || status != "active" {
		t.Fatalf("migrated role/status = %q/%q, want Agen/active", role, status)
	}

	var version int
	if err := db.QueryRowContext(ctx, `SELECT MAX(version) FROM schema_migrations`).Scan(&version); err != nil {
		t.Fatalf("read schema version: %v", err)
	}
	if version != schemaVersion {
		t.Fatalf("schema version = %d, want %d", version, schemaVersion)
	}
}
