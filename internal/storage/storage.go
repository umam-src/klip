package storage

import (
	"context"
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

const schemaVersion = 11

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
    role TEXT NOT NULL DEFAULT 'Agen',
    description TEXT NOT NULL DEFAULT '',
    provider_id TEXT NOT NULL DEFAULT '',
    model_id TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'active',
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

CREATE TABLE IF NOT EXISTS goal (
    id TEXT PRIMARY KEY,
    ruang_id TEXT NOT NULL REFERENCES ruang(id) ON DELETE CASCADE,
    parent_goal_id TEXT REFERENCES goal(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_goal_ruang ON goal(ruang_id);
CREATE INDEX IF NOT EXISTS idx_goal_parent ON goal(parent_goal_id);

CREATE TABLE IF NOT EXISTS proyek (
    id TEXT PRIMARY KEY,
    ruang_id TEXT NOT NULL REFERENCES ruang(id) ON DELETE CASCADE,
    goal_id TEXT NOT NULL REFERENCES goal(id) ON DELETE RESTRICT,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_proyek_ruang ON proyek(ruang_id);
CREATE INDEX IF NOT EXISTS idx_proyek_goal ON proyek(goal_id);

CREATE TABLE IF NOT EXISTS tugas (
    id TEXT PRIMARY KEY,
    ruang_id TEXT NOT NULL REFERENCES ruang(id) ON DELETE CASCADE,
    proyek_id TEXT NOT NULL REFERENCES proyek(id) ON DELETE CASCADE,
    parent_id TEXT REFERENCES tugas(id) ON DELETE SET NULL,
    title TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tugas_ruang ON tugas(ruang_id);
CREATE INDEX IF NOT EXISTS idx_tugas_proyek ON tugas(proyek_id);
CREATE INDEX IF NOT EXISTS idx_tugas_parent ON tugas(parent_id);

CREATE TABLE IF NOT EXISTS penugasan (
    id TEXT PRIMARY KEY,
    tugas_id TEXT NOT NULL REFERENCES tugas(id) ON DELETE CASCADE,
    agen_id TEXT NOT NULL REFERENCES agen(id) ON DELETE RESTRICT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    UNIQUE (tugas_id, agen_id)
);
CREATE INDEX IF NOT EXISTS idx_penugasan_tugas ON penugasan(tugas_id);
CREATE INDEX IF NOT EXISTS idx_penugasan_agen ON penugasan(agen_id);

CREATE TABLE IF NOT EXISTS eksekusi (
    id TEXT PRIMARY KEY,
    ruang_id TEXT NOT NULL REFERENCES ruang(id) ON DELETE CASCADE,
    proyek_id TEXT NOT NULL REFERENCES proyek(id) ON DELETE CASCADE,
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
CREATE INDEX IF NOT EXISTS idx_eksekusi_ruang ON eksekusi(ruang_id);
CREATE INDEX IF NOT EXISTS idx_eksekusi_proyek ON eksekusi(proyek_id);
CREATE INDEX IF NOT EXISTS idx_eksekusi_tugas ON eksekusi(tugas_id);
CREATE INDEX IF NOT EXISTS idx_eksekusi_agen ON eksekusi(agen_id);
CREATE INDEX IF NOT EXISTS idx_eksekusi_status ON eksekusi(status);

CREATE TABLE IF NOT EXISTS hasil (
    id TEXT PRIMARY KEY,
    ruang_id TEXT NOT NULL REFERENCES ruang(id) ON DELETE CASCADE,
    execution_id TEXT REFERENCES eksekusi(id) ON DELETE SET NULL,
    proyek_id TEXT NOT NULL REFERENCES proyek(id) ON DELETE CASCADE,
    tugas_id TEXT REFERENCES tugas(id) ON DELETE SET NULL,
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_hasil_ruang ON hasil(ruang_id);
CREATE INDEX IF NOT EXISTS idx_hasil_execution ON hasil(execution_id);
CREATE INDEX IF NOT EXISTS idx_hasil_proyek ON hasil(proyek_id);
CREATE INDEX IF NOT EXISTS idx_hasil_tugas ON hasil(tugas_id);

CREATE TABLE IF NOT EXISTS peristiwa (
    id TEXT PRIMARY KEY,
    ruang_id TEXT NOT NULL REFERENCES ruang(id) ON DELETE CASCADE,
    execution_id TEXT NOT NULL REFERENCES eksekusi(id) ON DELETE CASCADE,
    proyek_id TEXT NOT NULL REFERENCES proyek(id) ON DELETE CASCADE,
    tugas_id TEXT REFERENCES tugas(id) ON DELETE SET NULL,
    agen_id TEXT REFERENCES agen(id) ON DELETE RESTRICT,
    type TEXT NOT NULL,
    message TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_peristiwa_execution ON peristiwa(execution_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_peristiwa_proyek ON peristiwa(proyek_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_peristiwa_tugas ON peristiwa(tugas_id);

CREATE TABLE IF NOT EXISTS komentar (
    id TEXT PRIMARY KEY,
    ruang_id TEXT NOT NULL REFERENCES ruang(id) ON DELETE CASCADE,
    proyek_id TEXT NOT NULL REFERENCES proyek(id) ON DELETE CASCADE,
    tugas_id TEXT REFERENCES tugas(id) ON DELETE CASCADE,
    parent_id TEXT REFERENCES komentar(id) ON DELETE CASCADE,
    body TEXT NOT NULL,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_komentar_proyek_created ON komentar(proyek_id, created_at, id);
CREATE INDEX IF NOT EXISTS idx_komentar_tugas_created ON komentar(tugas_id, created_at, id);
CREATE INDEX IF NOT EXISTS idx_komentar_parent ON komentar(parent_id);

CREATE TABLE IF NOT EXISTS persetujuan (
    id TEXT PRIMARY KEY,
    ruang_id TEXT NOT NULL REFERENCES ruang(id) ON DELETE CASCADE,
    proyek_id TEXT NOT NULL REFERENCES proyek(id) ON DELETE CASCADE,
    tugas_id TEXT REFERENCES tugas(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    decided_at TEXT
);
CREATE INDEX IF NOT EXISTS idx_persetujuan_proyek ON persetujuan(proyek_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_persetujuan_tugas ON persetujuan(tugas_id, created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_persetujuan_status ON persetujuan(status);

CREATE TABLE IF NOT EXISTS jadwal (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    ruang_id TEXT NOT NULL REFERENCES ruang(id) ON DELETE CASCADE,
    proyek_id TEXT NOT NULL REFERENCES proyek(id) ON DELETE CASCADE,
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
CREATE INDEX IF NOT EXISTS idx_jadwal_due ON jadwal(status, next_run_at, id);

CREATE TABLE IF NOT EXISTS jadwal_eksekusi (
    id TEXT PRIMARY KEY,
    jadwal_id TEXT NOT NULL REFERENCES jadwal(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    attempt INTEGER NOT NULL,
    started_at TEXT NOT NULL,
    finished_at TEXT,
    error TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_jadwal_eksekusi_jadwal ON jadwal_eksekusi(jadwal_id, started_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_jadwal_eksekusi_status ON jadwal_eksekusi(status);

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
		return nil, fmt.Errorf("ping database: %w", err)
	}
	if err := migrate(ctx, db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func configure(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL; PRAGMA busy_timeout = 5000;`); err != nil {
		return fmt.Errorf("konfigurasi sqlite: %w", err)
	}
	return nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("buat skema database: %w", err)
	}
	var version int
	if err := db.QueryRowContext(ctx, `SELECT COALESCE(MAX(version), 0) FROM schema_migrations`).Scan(&version); err != nil {
		return fmt.Errorf("baca versi skema: %w", err)
	}
	if version > schemaVersion {
		return fmt.Errorf("versi database %d lebih baru dari aplikasi %d", version, schemaVersion)
	}
	if version < schemaVersion {
		if _, err := db.ExecContext(ctx, `INSERT INTO schema_migrations (version) VALUES (?)`, schemaVersion); err != nil {
			return fmt.Errorf("catat versi skema: %w", err)
		}
	}
	return nil
}
