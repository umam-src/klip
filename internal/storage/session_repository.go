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

type SesiRepository interface {
	CreateSesi(ctx context.Context, sesi domain.Sesi) error
	GetSesi(ctx context.Context, id domain.ID) (domain.Sesi, error)
	FinishSesi(ctx context.Context, id domain.ID, status domain.Status, finishedAt time.Time) error
}

func (r *Repository) CreateSesi(ctx context.Context, sesi domain.Sesi) error {
	if strings.TrimSpace(string(sesi.ID)) == "" ||
		strings.TrimSpace(string(sesi.PekerjaanID)) == "" ||
		strings.TrimSpace(string(sesi.AgenID)) == "" ||
		sesi.Status != domain.StatusRunning {
		return fmt.Errorf("sesi: %w", ErrInvalid)
	}
	started := sesi.StartedAt
	if started.IsZero() {
		started = time.Now().UTC()
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO sesi (id, pekerjaan_id, agen_id, status, started_at, finished_at)
		VALUES (?, ?, ?, ?, ?, NULL)`,
		sesi.ID, sesi.PekerjaanID, sesi.AgenID, sesi.Status,
		started.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("buat sesi: %w", err)
	}
	return nil
}

func (r *Repository) GetSesi(ctx context.Context, id domain.ID) (domain.Sesi, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Sesi{}, fmt.Errorf("sesi: %w", ErrInvalid)
	}
	var sesi domain.Sesi
	var started, finished sql.NullString
	err := r.db.QueryRowContext(ctx, `
		SELECT id, pekerjaan_id, agen_id, status, started_at, finished_at
		FROM sesi WHERE id = ?`, id).Scan(
		&sesi.ID, &sesi.PekerjaanID, &sesi.AgenID, &sesi.Status, &started, &finished,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Sesi{}, ErrNotFound
	}
	if err != nil {
		return domain.Sesi{}, fmt.Errorf("ambil sesi: %w", err)
	}
	var parseErr error
	if sesi.StartedAt, parseErr = parseTime(started.String); parseErr != nil {
		return domain.Sesi{}, parseErr
	}
	if finished.Valid {
		value, err := parseTime(finished.String)
		if err != nil {
			return domain.Sesi{}, err
		}
		sesi.FinishedAt = &value
	}
	return sesi, nil
}

func (r *Repository) FinishSesi(ctx context.Context, id domain.ID, status domain.Status, finishedAt time.Time) error {
	if strings.TrimSpace(string(id)) == "" || !isTerminalStatus(status) {
		return fmt.Errorf("sesi: %w", ErrInvalid)
	}
	var current domain.Status
	if err := r.db.QueryRowContext(ctx, `SELECT status FROM sesi WHERE id = ?`, id).Scan(&current); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("baca status sesi: %w", err)
	}
	if current != domain.StatusRunning || !current.CanTransitionTo(status) {
		return fmt.Errorf("sesi: transisi status %q ke %q tidak diizinkan: %w", current, status, ErrInvalid)
	}
	if finishedAt.IsZero() {
		finishedAt = time.Now().UTC()
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE sesi SET status = ?, finished_at = ? WHERE id = ?`,
		status, finishedAt.UTC().Format(time.RFC3339Nano), id,
	)
	if err != nil {
		return fmt.Errorf("selesaikan sesi: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func isTerminalStatus(status domain.Status) bool {
	switch status {
	case domain.StatusCompleted, domain.StatusFailed, domain.StatusCancelled:
		return true
	default:
		return false
	}
}
