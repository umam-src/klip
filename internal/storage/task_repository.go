package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

type TugasLifecycleRepository interface {
	UpdateTugasStatus(ctx context.Context, id domain.ID, status domain.Status, updatedAt time.Time) error
}

func (r *Repository) UpdateTugasStatus(ctx context.Context, id domain.ID, status domain.Status, updatedAt time.Time) error {
	if strings.TrimSpace(string(id)) == "" || strings.TrimSpace(string(status)) == "" || !isKnownStatus(status) {
		return fmt.Errorf("tugas: %w", ErrInvalid)
	}

	var current domain.Status
	if err := r.db.QueryRowContext(ctx, `SELECT status FROM tugas WHERE id = ?`, id).Scan(&current); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("baca status tugas: %w", err)
	}
	if !current.CanTransitionTo(status) {
		return fmt.Errorf("tugas: transisi status %q ke %q tidak diizinkan: %w", current, status, ErrInvalid)
	}

	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE tugas SET status = ?, updated_at = ? WHERE id = ?`,
		status, updatedAt.UTC().Format(time.RFC3339Nano), id,
	)
	if err != nil {
		return fmt.Errorf("ubah status tugas: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func isKnownStatus(status domain.Status) bool {
	switch status {
	case domain.StatusDraft, domain.StatusReady, domain.StatusRunning, domain.StatusWaiting, domain.StatusBlocked, domain.StatusCompleted, domain.StatusFailed, domain.StatusCancelled:
		return true
	default:
		return false
	}
}
