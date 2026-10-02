package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

type ExecutionStart struct {
	Sesi domain.Sesi
	Run  domain.Run
}

// StartExecution creates the running session and run, and moves the task to
// running, in one SQLite transaction. The transaction ends before the process
// is started so it never stays open while external work is running.
func (r *Repository) StartExecution(ctx context.Context, sesi domain.Sesi, run domain.Run) (ExecutionStart, error) {
	if strings.TrimSpace(string(sesi.ID)) == "" ||
		strings.TrimSpace(string(sesi.PekerjaanID)) == "" ||
		strings.TrimSpace(string(sesi.AgenID)) == "" ||
		sesi.Status != domain.StatusRunning ||
		strings.TrimSpace(string(run.ID)) == "" ||
		strings.TrimSpace(string(run.PekerjaanID)) == "" ||
		strings.TrimSpace(string(run.AgenID)) == "" ||
		strings.TrimSpace(run.Program) == "" ||
		run.Status != domain.StatusRunning {
		return ExecutionStart{}, fmt.Errorf("execution: %w", ErrInvalid)
	}
	if sesi.PekerjaanID != run.PekerjaanID || sesi.AgenID != run.AgenID {
		return ExecutionStart{}, fmt.Errorf("execution: sesi dan run tidak konsisten: %w", ErrInvalid)
	}
	if run.TugasID == nil {
		return r.startExecutionTx(ctx, sesi, run, nil)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ExecutionStart{}, fmt.Errorf("mulai transaksi execution: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var taskPekerjaanID domain.ID
	var current domain.Status
	if err := tx.QueryRow(`SELECT pekerjaan_id, status FROM tugas WHERE id = ?`, *run.TugasID).Scan(&taskPekerjaanID, &current); err != nil {
		if err == sql.ErrNoRows {
			return ExecutionStart{}, ErrNotFound
		}
		return ExecutionStart{}, fmt.Errorf("baca tugas execution: %w", err)
	}
	if taskPekerjaanID != run.PekerjaanID {
		return ExecutionStart{}, fmt.Errorf("execution: tugas tidak termasuk pekerjaan: %w", ErrInvalid)
	}
	if !current.CanTransitionTo(domain.StatusRunning) {
		return ExecutionStart{}, fmt.Errorf("execution: tugas berstatus %q tidak dapat dijalankan: %w", current, ErrInvalid)
	}

	started := sesi.StartedAt
	if started.IsZero() {
		started = time.Now().UTC()
	}
	if err := insertSesiTx(tx, sesi, started); err != nil {
		return ExecutionStart{}, err
	}
	if _, err := tx.Exec(`UPDATE tugas SET status = ?, updated_at = ? WHERE id = ?`, domain.StatusRunning, started.UTC().Format(time.RFC3339Nano), *run.TugasID); err != nil {
		return ExecutionStart{}, fmt.Errorf("mulai tugas: %w", err)
	}
	if err := insertRunTx(tx, run); err != nil {
		return ExecutionStart{}, err
	}
	if err := tx.Commit(); err != nil {
		return ExecutionStart{}, fmt.Errorf("commit execution: %w", err)
	}
	sesi.StartedAt = started
	run.StartedAt = started
	return ExecutionStart{Sesi: sesi, Run: run}, nil
}

func (r *Repository) startExecutionTx(ctx context.Context, sesi domain.Sesi, run domain.Run, _ *domain.ID) (ExecutionStart, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return ExecutionStart{}, fmt.Errorf("mulai transaksi execution: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	started := sesi.StartedAt
	if started.IsZero() {
		started = time.Now().UTC()
	}
	if err := insertSesiTx(tx, sesi, started); err != nil {
		return ExecutionStart{}, err
	}
	if err := insertRunTx(tx, run); err != nil {
		return ExecutionStart{}, err
	}
	if err := tx.Commit(); err != nil {
		return ExecutionStart{}, fmt.Errorf("commit execution: %w", err)
	}
	sesi.StartedAt = started
	run.StartedAt = started
	return ExecutionStart{Sesi: sesi, Run: run}, nil
}

func insertSesiTx(tx *sql.Tx, sesi domain.Sesi, started time.Time) error {
	if _, err := tx.Exec(`
		INSERT INTO sesi (id, pekerjaan_id, agen_id, status, started_at, finished_at)
		VALUES (?, ?, ?, ?, ?, NULL)`,
		sesi.ID, sesi.PekerjaanID, sesi.AgenID, sesi.Status,
		started.UTC().Format(time.RFC3339Nano),
	); err != nil {
		return fmt.Errorf("buat sesi: %w", err)
	}
	return nil
}

func insertRunTx(tx *sql.Tx, run domain.Run) error {
	arguments, err := json.Marshal(run.Arguments)
	if err != nil {
		return fmt.Errorf("run: serialisasi argumen: %w", err)
	}
	started := run.StartedAt
	if started.IsZero() {
		started = time.Now().UTC()
	}
	var tugasID any
	if run.TugasID != nil {
		tugasID = string(*run.TugasID)
	}
	if _, err := tx.Exec(`
		INSERT INTO run (
			id, pekerjaan_id, tugas_id, agen_id, status, program, arguments,
			exit_code, stdout, stderr, started_at, finished_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, NULL, ?, ?, ?, NULL)`,
		run.ID, run.PekerjaanID, tugasID, run.AgenID, run.Status, run.Program,
		string(arguments), run.Stdout, run.Stderr, started.UTC().Format(time.RFC3339Nano),
	); err != nil {
		return fmt.Errorf("buat run: %w", err)
	}
	return nil
}
