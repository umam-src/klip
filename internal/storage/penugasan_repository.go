package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/umam-src/klip/internal/domain"
)

// PenugasanRepository menyediakan batas penyimpanan native untuk hubungan
// Tugas-Agen. Keunikan pasangan tugas dan agen dijaga oleh basis data.
type PenugasanRepository interface {
	CreatePenugasan(ctx context.Context, penugasan domain.Penugasan) error
	GetPenugasan(ctx context.Context, id domain.ID) (domain.Penugasan, error)
	ListPenugasanByTugas(ctx context.Context, tugasID domain.ID) ([]domain.Penugasan, error)
	ListPenugasanByAgen(ctx context.Context, agenID domain.ID) ([]domain.Penugasan, error)
}

func (r *Repository) CreatePenugasan(ctx context.Context, penugasan domain.Penugasan) error {
	if strings.TrimSpace(string(penugasan.ID)) == "" ||
		strings.TrimSpace(string(penugasan.TugasID)) == "" ||
		strings.TrimSpace(string(penugasan.AgenID)) == "" {
		return fmt.Errorf("penugasan: %w", ErrInvalid)
	}

	var tugasRuangID, agenRuangID string
	if err := r.db.QueryRowContext(ctx, `SELECT ruang_id FROM tugas WHERE id = ?`, penugasan.TugasID).Scan(&tugasRuangID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("penugasan: tugas tidak ditemukan: %w", ErrInvalid)
		}
		return fmt.Errorf("cek tugas penugasan: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT ruang_id FROM agen WHERE id = ?`, penugasan.AgenID).Scan(&agenRuangID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("penugasan: agen tidak ditemukan: %w", ErrInvalid)
		}
		return fmt.Errorf("cek agen penugasan: %w", err)
	}
	if tugasRuangID != agenRuangID {
		return fmt.Errorf("penugasan: tugas dan agen harus berada di ruang yang sama: %w", ErrInvalid)
	}

	created, updated := timestamps(penugasan.CreatedAt, penugasan.UpdatedAt)
	_, err := r.db.ExecContext(ctx, `INSERT INTO penugasan (id, tugas_id, agen_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, penugasan.ID, penugasan.TugasID, penugasan.AgenID, created, updated)
	if err != nil {
		return fmt.Errorf("buat penugasan: %w", err)
	}
	return nil
}

func (r *Repository) GetPenugasan(ctx context.Context, id domain.ID) (domain.Penugasan, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Penugasan{}, fmt.Errorf("penugasan: %w", ErrInvalid)
	}

	var penugasan domain.Penugasan
	var created, updated string
	err := r.db.QueryRowContext(ctx, `SELECT id, tugas_id, agen_id, created_at, updated_at FROM penugasan WHERE id = ?`, id).Scan(&penugasan.ID, &penugasan.TugasID, &penugasan.AgenID, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Penugasan{}, ErrNotFound
	}
	if err != nil {
		return domain.Penugasan{}, fmt.Errorf("ambil penugasan: %w", err)
	}

	var parseErr error
	if penugasan.CreatedAt, parseErr = parseTime(created); parseErr != nil {
		return domain.Penugasan{}, parseErr
	}
	if penugasan.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
		return domain.Penugasan{}, parseErr
	}
	return penugasan, nil
}

func (r *Repository) ListPenugasanByTugas(ctx context.Context, tugasID domain.ID) ([]domain.Penugasan, error) {
	return r.listPenugasan(ctx, `tugas_id`, tugasID)
}

func (r *Repository) ListPenugasanByAgen(ctx context.Context, agenID domain.ID) ([]domain.Penugasan, error) {
	return r.listPenugasan(ctx, `agen_id`, agenID)
}

func (r *Repository) listPenugasan(ctx context.Context, column string, id domain.ID) ([]domain.Penugasan, error) {
	if strings.TrimSpace(string(id)) == "" {
		return nil, fmt.Errorf("penugasan: id wajib diisi: %w", ErrInvalid)
	}

	query := `SELECT id, tugas_id, agen_id, created_at, updated_at FROM penugasan WHERE ` + column + ` = ? ORDER BY created_at, id`
	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("daftar penugasan: %w", err)
	}
	defer rows.Close()

	var result []domain.Penugasan
	for rows.Next() {
		var penugasan domain.Penugasan
		var created, updated string
		if err := rows.Scan(&penugasan.ID, &penugasan.TugasID, &penugasan.AgenID, &created, &updated); err != nil {
			return nil, fmt.Errorf("baca penugasan: %w", err)
		}
		var parseErr error
		if penugasan.CreatedAt, parseErr = parseTime(created); parseErr != nil {
			return nil, parseErr
		}
		if penugasan.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
			return nil, parseErr
		}
		result = append(result, penugasan)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar penugasan: %w", err)
	}
	return result, nil
}
