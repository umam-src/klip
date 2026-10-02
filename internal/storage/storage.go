package storage

import (
    "context"
    "database/sql"
    "fmt"

    _ "modernc.org/sqlite"
)

const schemaVersion = 1

const schema = `
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA busy_timeout = 5000;

CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER PRIMARY KEY,
    applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ruang (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS agen (
    id TEXT PRIMARY KEY,
    ruang_id TEXT NOT NULL REFERENCES ruang(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    provider_id TEXT NOT NULL DEFAULT '',
    model_id TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_agen_ruang ON agen(ruang_id);

CREATE TABLE IF NOT EXISTS sasaran (
    id TEXT PRIMARY KEY,
    ruang_id TEXT NOT NULL REFERENCES ruang(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sasaran_ruang ON sasaran(ruang_id);

CREATE TABLE IF NOT EXISTS pekerjaan (
    id TEXT PRIMARY KEY,
    ruang_id TEXT NOT NULL REFERENCES ruang(id) ON DELETE CASCADE,
    sasaran_id TEXT REFERENCES sasaran(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_pekerjaan_ruang ON pekerjaan(ruang_id);
CREATE INDEX IF NOT EXISTS idx_pekerjaan_sasaran ON pekerjaan(sasaran_id);

CREATE TABLE IF NOT EXISTS tugas (
    id TEXT PRIMARY KEY,
    pekerjaan_id TEXT NOT NULL REFERENCES pekerjaan(id) ON DELETE CASCADE,
    parent_id TEXT REFERENCES tugas(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL,
    position INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tugas_pekerjaan ON tugas(pekerjaan_id);

CREATE TABLE IF NOT EXISTS sesi (
    id TEXT PRIMARY KEY,
    pekerjaan_id TEXT NOT NULL REFERENCES pekerjaan(id) ON DELETE CASCADE,
    agen_id TEXT NOT NULL REFERENCES agen(id) ON DELETE RESTRICT,
    status TEXT NOT NULL,
    started_at TEXT NOT NULL,
    finished_at TEXT
);
CREATE INDEX IF NOT EXISTS idx_sesi_pekerjaan ON sesi(pekerjaan_id);
CREATE INDEX IF NOT EXISTS idx_sesi_agen ON sesi(agen_id);

CREATE TABLE IF NOT EXISTS hasil (
    id TEXT PRIMARY KEY,
    pekerjaan_id TEXT NOT NULL REFERENCES pekerjaan(id) ON DELETE CASCADE,
    tugas_id TEXT REFERENCES tugas(id) ON DELETE SET NULL,
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_hasil_pekerjaan ON hasil(pekerjaan_id);
CREATE INDEX IF NOT EXISTS idx_hasil_tugas ON hasil(tugas_id);
`

// Open membuka database lokal dan memastikan skema minimum Klip tersedia.
func Open(ctx context.Context, path string) (*sql.DB, error) {
    db, err := sql.Open("sqlite", path)
    if err != nil {
        return nil, fmt.Errorf("buka database: %w", err)
    }

    db.SetMaxOpenConns(1)
    db.SetMaxIdleConns(1)

    if err := db.PingContext(ctx); err != nil {
        _ = db.Close()
        return nil, fmt.Errorf("hubungkan database: %w", err)
    }

    if err := migrate(ctx, db); err != nil {
        _ = db.Close()
        return nil, err
    }

    return db, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
    tx, err := db.BeginTx(ctx, nil)
    if err != nil {
        return fmt.Errorf("mulai migrasi: %w", err)
    }
    defer tx.Rollback()

    if _, err := tx.ExecContext(ctx, schema); err != nil {
        return fmt.Errorf("buat skema: %w", err)
    }

    var version int
    if err := tx.QueryRowContext(ctx, "SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version); err != nil {
        return fmt.Errorf("baca versi skema: %w", err)
    }

    if version < schemaVersion {
        if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations(version) VALUES (?)", schemaVersion); err != nil {
            return fmt.Errorf("catat migrasi: %w", err)
        }
    }

    if err := tx.Commit(); err != nil {
        return fmt.Errorf("simpan migrasi: %w", err)
    }
    return nil
}
