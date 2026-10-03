package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const schemaVersion = 9

const schema = `
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
    parent_id TEXT REFERENCES agen(id) ON DELETE SET NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    provider_id TEXT NOT NULL DEFAULT '',
    model_id TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_agen_ruang ON agen(ruang_id);
CREATE INDEX IF NOT EXISTS idx_agen_parent ON agen(parent_id);

CREATE TABLE IF NOT EXISTS agen_skill (
    agen_id TEXT NOT NULL REFERENCES agen(id) ON DELETE CASCADE,
    skill_name TEXT NOT NULL,
    PRIMARY KEY (agen_id, skill_name)
);
CREATE INDEX IF NOT EXISTS idx_agen_skill_name ON agen_skill(skill_name);

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

CREATE TABLE IF NOT EXISTS run (
    id TEXT PRIMARY KEY,
    pekerjaan_id TEXT NOT NULL REFERENCES pekerjaan(id) ON DELETE CASCADE,
    tugas_id TEXT REFERENCES tugas(id) ON DELETE SET NULL,
    agen_id TEXT NOT NULL REFERENCES agen(id) ON DELETE RESTRICT,
    status TEXT NOT NULL,
    program TEXT NOT NULL,
    arguments TEXT NOT NULL DEFAULT '[]',
    exit_code INTEGER,
    stdout TEXT NOT NULL DEFAULT '',
    stderr TEXT NOT NULL DEFAULT '',
    started_at TEXT NOT NULL,
    finished_at TEXT
);
CREATE INDEX IF NOT EXISTS idx_run_pekerjaan ON run(pekerjaan_id);
CREATE INDEX IF NOT EXISTS idx_run_tugas ON run(tugas_id);
CREATE INDEX IF NOT EXISTS idx_run_agen ON run(agen_id);
CREATE INDEX IF NOT EXISTS idx_run_status ON run(status);

CREATE TABLE IF NOT EXISTS event (
    id TEXT PRIMARY KEY,
    pekerjaan_id TEXT NOT NULL REFERENCES pekerjaan(id) ON DELETE CASCADE,
    tugas_id TEXT REFERENCES tugas(id) ON DELETE SET NULL,
    sesi_id TEXT REFERENCES sesi(id) ON DELETE CASCADE,
    run_id TEXT REFERENCES run(id) ON DELETE CASCADE,
    agen_id TEXT REFERENCES agen(id) ON DELETE RESTRICT,
    type TEXT NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_event_pekerjaan_created ON event(pekerjaan_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_event_sesi ON event(sesi_id);
CREATE INDEX IF NOT EXISTS idx_event_run ON event(run_id);

CREATE TABLE IF NOT EXISTS komentar (
    id TEXT PRIMARY KEY,
    pekerjaan_id TEXT NOT NULL REFERENCES pekerjaan(id) ON DELETE CASCADE,
    tugas_id TEXT REFERENCES tugas(id) ON DELETE CASCADE,
    parent_id TEXT REFERENCES komentar(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_komentar_pekerjaan_created ON komentar(pekerjaan_id, created_at, id);
CREATE INDEX IF NOT EXISTS idx_komentar_tugas_created ON komentar(tugas_id, created_at, id);
CREATE INDEX IF NOT EXISTS idx_komentar_parent ON komentar(parent_id);

CREATE TABLE IF NOT EXISTS approval (
    id TEXT PRIMARY KEY,
    pekerjaan_id TEXT NOT NULL REFERENCES pekerjaan(id) ON DELETE CASCADE,
    tugas_id TEXT REFERENCES tugas(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    decided_at TEXT
);
CREATE INDEX IF NOT EXISTS idx_approval_pekerjaan ON approval(pekerjaan_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_approval_tugas ON approval(tugas_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_approval_status ON approval(status);

CREATE TABLE IF NOT EXISTS schedule (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    pekerjaan_id TEXT NOT NULL REFERENCES pekerjaan(id) ON DELETE CASCADE,
    tugas_id TEXT REFERENCES tugas(id) ON DELETE SET NULL,
    agen_id TEXT NOT NULL REFERENCES agen(id) ON DELETE RESTRICT,
    program TEXT NOT NULL,
    arguments TEXT NOT NULL DEFAULT '[]',
    interval_seconds INTEGER NOT NULL,
    next_run_at TEXT NOT NULL,
    status TEXT NOT NULL,
    retry_limit INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_schedule_due ON schedule(status, next_run_at, id);

CREATE TABLE IF NOT EXISTS schedule_run (
    id TEXT PRIMARY KEY,
    schedule_id TEXT NOT NULL REFERENCES schedule(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    attempt INTEGER NOT NULL,
    started_at TEXT NOT NULL,
    finished_at TEXT,
    error TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_schedule_run_schedule ON schedule_run(schedule_id, started_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_schedule_run_status ON schedule_run(status);

CREATE TABLE IF NOT EXISTS pengaturan_ai (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    provider TEXT NOT NULL,
    base_url TEXT NOT NULL,
    model TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
`

// Open membuka database lokal dan memastikan skema minimum Klip tersedia.
func Open(ctx context.Context, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("buka database: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if err := configure(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
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

func configure(ctx context.Context, db *sql.DB) error {
	for _, statement := range []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
	} {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("atur SQLite: %w", err)
		}
	}
	return nil
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
	if err := ensureAgenParentColumn(ctx, tx); err != nil {
		return err
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

func ensureAgenParentColumn(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, "PRAGMA table_info(agen)")
	if err != nil {
		return fmt.Errorf("baca kolom agen: %w", err)
	}
	defer rows.Close()

	var found bool
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull, primaryKey int
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			return fmt.Errorf("baca metadata agen: %w", err)
		}
		if name == "parent_id" {
			found = true
			break
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("baca metadata agen: %w", err)
	}
	if !found {
		if _, err := tx.ExecContext(ctx, "ALTER TABLE agen ADD COLUMN parent_id TEXT REFERENCES agen(id) ON DELETE SET NULL"); err != nil {
			return fmt.Errorf("migrasi hierarki agen: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, "CREATE INDEX IF NOT EXISTS idx_agen_parent ON agen(parent_id)"); err != nil {
		return fmt.Errorf("indeks hierarki agen: %w", err)
	}
	return nil
}
