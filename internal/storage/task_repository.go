package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

type TugasLifecycleRepository interface {
	UpdateTugasStatus(ctx context.Context, id domain.ID, status domain.Status, updatedAt time.Time) error
}

func (r *Repository) UpdateTugasStatus(ctx context.Context, id domain.ID, status domain.Status, updatedAt time.Time) error {
	if strings.TrimSpace(string(id)) == "" || strings.TrimSpace(string(status)) == "" {
		return fmt.Errorf("tugas: %w", ErrInvalid)
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
