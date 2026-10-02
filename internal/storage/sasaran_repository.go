package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/umam-src/klip/internal/domain"
)

func (r *Repository) CreateSasaran(ctx context.Context, sasaran domain.Sasaran) error {
	if err := validateIDName(sasaran.ID, sasaran.Title); err != nil {
		return fmt.Errorf("sasaran: %w", err)
	}
	if strings.TrimSpace(string(sasaran.RuangID)) == "" {
		return fmt.Errorf("sasaran: ruang wajib diisi: %w", ErrInvalid)
	}
	status := sasaran.Status
	if status == "" {
		status = domain.StatusDraft
	}
	created, updated := timestamps(sasaran.CreatedAt, sasaran.UpdatedAt)
	_, err := r.db.ExecContext(ctx, `INSERT INTO sasaran (id, ruang_id, title, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`, sasaran.ID, sasaran.RuangID, sasaran.Title, status, created, updated)
	if err != nil {
		return fmt.Errorf("buat sasaran: %w", err)
	}
	return nil
}

func (r *Repository) GetSasaran(ctx context.Context, id domain.ID) (domain.Sasaran, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Sasaran{}, fmt.Errorf("sasaran: %w", ErrInvalid)
	}
	var sasaran domain.Sasaran
	var created, updated string
	err := r.db.QueryRowContext(ctx, `SELECT id, ruang_id, title, status, created_at, updated_at FROM sasaran WHERE id = ?`, id).Scan(&sasaran.ID, &sasaran.RuangID, &sasaran.Title, &sasaran.Status, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Sasaran{}, ErrNotFound
	}
	if err != nil {
		return domain.Sasaran{}, fmt.Errorf("ambil sasaran: %w", err)
	}
	var parseErr error
	if sasaran.CreatedAt, parseErr = parseTime(created); parseErr != nil {
		return domain.Sasaran{}, parseErr
	}
	if sasaran.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
		return domain.Sasaran{}, parseErr
	}
	return sasaran, nil
}

func (r *Repository) ListSasaranByRuang(ctx context.Context, ruangID domain.ID) ([]domain.Sasaran, error) {
	if strings.TrimSpace(string(ruangID)) == "" {
		return nil, fmt.Errorf("sasaran: ruang wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, ruang_id, title, status, created_at, updated_at FROM sasaran WHERE ruang_id = ? ORDER BY created_at, id`, ruangID)
	if err != nil {
		return nil, fmt.Errorf("daftar sasaran: %w", err)
	}
	defer rows.Close()

	var result []domain.Sasaran
	for rows.Next() {
		var sasaran domain.Sasaran
		var created, updated string
		if err := rows.Scan(&sasaran.ID, &sasaran.RuangID, &sasaran.Title, &sasaran.Status, &created, &updated); err != nil {
			return nil, fmt.Errorf("baca sasaran: %w", err)
		}
		var parseErr error
		if sasaran.CreatedAt, parseErr = parseTime(created); parseErr != nil {
			return nil, parseErr
		}
		if sasaran.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
			return nil, parseErr
		}
		result = append(result, sasaran)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar sasaran: %w", err)
	}
	return result, nil
}
