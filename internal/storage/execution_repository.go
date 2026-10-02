package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

type ExecutionRepository interface {
	FinalizeExecution(ctx context.Context, runID, sesiID domain.ID, tugasID *domain.ID, status domain.Status, exitCode *int, stdout, stderr string, finishedAt time.Time) error
}

// FinalizeExecution changes all execution state and records the terminal event
// in one SQLite transaction. This prevents a partial finalization where Run,
// Sesi, Tugas, and the runtime event disagree.
func (r *Repository) FinalizeExecution(ctx context.Context, runID, sesiID domain.ID, tugasID *domain.ID, status domain.Status, exitCode *int, stdout, stderr string, finishedAt time.Time) error {
	if strings.TrimSpace(string(runID)) == "" || strings.TrimSpace(string(sesiID)) == "" || !isTerminalStatus(status) {
		return fmt.Errorf("execution: %w", ErrInvalid)
	}
	if finishedAt.IsZero() {
		finishedAt = time.Now().UTC()
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mulai transaksi execution: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := finishRunTx(tx, runID, status, exitCode, stdout, stderr, finishedAt); err != nil {
		return err
	}
	if err := finishSesiTx(tx, sesiID, status, finishedAt); err != nil {
		return err
	}
	if tugasID != nil {
		if err := finishTugasTx(tx, *tugasID, status, finishedAt); err != nil {
			return err
		}
	}

	var pekerjaanID, agenID domain.ID
	var eventTugasID sql.NullString
	if err := tx.QueryRow(`SELECT pekerjaan_id, agen_id, tugas_id FROM run WHERE id = ?`, runID).Scan(&pekerjaanID, &agenID, &eventTugasID); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return fmt.Errorf("baca konteks event execution: %w", err)
	}
	var eventTaskID *domain.ID
	if eventTugasID.Valid {
		value := domain.ID(eventTugasID.String)
		eventTaskID = &value
	}
	eventType := domain.EventTypeForStatus(status)
	if eventType == "" {
		return fmt.Errorf("execution: status event tidak dikenal: %q: %w", status, ErrInvalid)
	}
	if err := appendEvent(ctx, tx, domain.Event{
		PekerjaanID: pekerjaanID,
		TugasID:     eventTaskID,
		SesiID:      idPtr(sesiID),
		RunID:       idPtr(runID),
		AgenID:      idPtr(agenID),
		Type:        eventType,
		Message:     "Eksekusi selesai dengan status " + string(status),
		CreatedAt:   finishedAt,
	}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit execution: %w", err)
	}
	return nil
}

func finishRunTx(tx *sql.Tx, id domain.ID, status domain.Status, exitCode *int, stdout, stderr string, finishedAt time.Time) error {
	var current domain.Status
	if err := tx.QueryRow(`SELECT status FROM run WHERE id = ?`, id).Scan(&current); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return fmt.Errorf("baca status run: %w", err)
	}
	if current != domain.StatusRunning || !current.CanTransitionTo(status) {
		return fmt.Errorf("run: transisi status %q ke %q tidak diizinkan: %w", current, status, ErrInvalid)
	}
	var exit any
	if exitCode != nil {
		exit = *exitCode
	}
	result, err := tx.Exec(`UPDATE run SET status = ?, exit_code = ?, stdout = ?, stderr = ?, finished_at = ? WHERE id = ?`, status, exit, stdout, stderr, finishedAt.UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return fmt.Errorf("selesaikan run: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func finishSesiTx(tx *sql.Tx, id domain.ID, status domain.Status, finishedAt time.Time) error {
	var current domain.Status
	if err := tx.QueryRow(`SELECT status FROM sesi WHERE id = ?`, id).Scan(&current); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return fmt.Errorf("baca status sesi: %w", err)
	}
	if current != domain.StatusRunning || !current.CanTransitionTo(status) {
		return fmt.Errorf("sesi: transisi status %q ke %q tidak diizinkan: %w", current, status, ErrInvalid)
	}
	result, err := tx.Exec(`UPDATE sesi SET status = ?, finished_at = ? WHERE id = ?`, status, finishedAt.UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return fmt.Errorf("selesaikan sesi: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func finishTugasTx(tx *sql.Tx, id domain.ID, status domain.Status, updatedAt time.Time) error {
	var current domain.Status
	if err := tx.QueryRow(`SELECT status FROM tugas WHERE id = ?`, id).Scan(&current); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return fmt.Errorf("baca status tugas: %w", err)
	}
	if current != domain.StatusRunning || !current.CanTransitionTo(status) {
		return fmt.Errorf("tugas: transisi status %q ke %q tidak diizinkan: %w", current, status, ErrInvalid)
	}
	result, err := tx.Exec(`UPDATE tugas SET status = ?, updated_at = ? WHERE id = ?`, status, updatedAt.UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return fmt.Errorf("selesaikan tugas: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}
