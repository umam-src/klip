package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/umam-src/klip/internal/domain"
)

// ProyekRepository provides the native Proyek persistence boundary.
type ProyekRepository interface {
	CreateProyek(ctx context.Context, proyek domain.Proyek) error
	GetProyek(ctx context.Context, id domain.ID) (domain.Proyek, error)
	ListProyekByRuang(ctx context.Context, ruangID domain.ID) ([]domain.Proyek, error)
}

func (r *Repository) CreateProyek(ctx context.Context, proyek domain.Proyek) error {
	if strings.TrimSpace(string(proyek.ID)) == "" || strings.TrimSpace(string(proyek.RuangID)) == "" || strings.TrimSpace(string(proyek.GoalID)) == "" || strings.TrimSpace(proyek.Title) == "" {
		return fmt.Errorf("proyek: %w", ErrInvalid)
	}

	var goalRuangID string
	err := r.db.QueryRowContext(ctx, `SELECT ruang_id FROM goal WHERE id = ?`, proyek.GoalID).Scan(&goalRuangID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("proyek: goal tidak ditemukan: %w", ErrInvalid)
	}
	if err != nil {
		return fmt.Errorf("cek goal proyek: %w", err)
	}
	if goalRuangID != string(proyek.RuangID) {
		return fmt.Errorf("proyek: goal tidak sesuai dengan ruang: %w", ErrInvalid)
	}

	created, updated := timestamps(proyek.CreatedAt, proyek.UpdatedAt)
	_, err = r.db.ExecContext(ctx, `INSERT INTO proyek (id, ruang_id, goal_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, proyek.ID, proyek.RuangID, proyek.GoalID, proyek.Title, proyek.Description, proyek.Status, created, updated)
	if err != nil {
		return fmt.Errorf("buat proyek: %w", err)
	}
	return nil
}

func (r *Repository) GetProyek(ctx context.Context, id domain.ID) (domain.Proyek, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Proyek{}, fmt.Errorf("proyek: %w", ErrInvalid)
	}

	var proyek domain.Proyek
	var created, updated string
	err := r.db.QueryRowContext(ctx, `SELECT id, ruang_id, goal_id, title, description, status, created_at, updated_at FROM proyek WHERE id = ?`, id).Scan(&proyek.ID, &proyek.RuangID, &proyek.GoalID, &proyek.Title, &proyek.Description, &proyek.Status, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Proyek{}, ErrNotFound
	}
	if err != nil {
		return domain.Proyek{}, fmt.Errorf("ambil proyek: %w", err)
	}

	var parseErr error
	if proyek.CreatedAt, parseErr = parseTime(created); parseErr != nil {
		return domain.Proyek{}, parseErr
	}
	if proyek.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
		return domain.Proyek{}, parseErr
	}
	return proyek, nil
}

func (r *Repository) ListProyekByRuang(ctx context.Context, ruangID domain.ID) ([]domain.Proyek, error) {
	if strings.TrimSpace(string(ruangID)) == "" {
		return nil, fmt.Errorf("proyek: ruang wajib diisi: %w", ErrInvalid)
	}

	rows, err := r.db.QueryContext(ctx, `SELECT id, ruang_id, goal_id, title, description, status, created_at, updated_at FROM proyek WHERE ruang_id = ? ORDER BY created_at, id`, ruangID)
	if err != nil {
		return nil, fmt.Errorf("daftar proyek: %w", err)
	}
	defer rows.Close()

	var result []domain.Proyek
	for rows.Next() {
		var proyek domain.Proyek
		var created, updated string
		if err := rows.Scan(&proyek.ID, &proyek.RuangID, &proyek.GoalID, &proyek.Title, &proyek.Description, &proyek.Status, &created, &updated); err != nil {
			return nil, fmt.Errorf("baca proyek: %w", err)
		}
		var parseErr error
		if proyek.CreatedAt, parseErr = parseTime(created); parseErr != nil {
			return nil, parseErr
		}
		if proyek.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
			return nil, parseErr
		}
		result = append(result, proyek)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar proyek: %w", err)
	}
	return result, nil
}
