package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/umam-src/klip/internal/domain"
)

// TugasNativeRepository provides the v1 Tugas persistence boundary while the
// legacy repository is migrated away from Pekerjaan.
type TugasNativeRepository interface {
	CreateTugasNative(ctx context.Context, tugas domain.Tugas) error
	GetTugasNative(ctx context.Context, id domain.ID) (domain.Tugas, error)
	ListTugasByProyek(ctx context.Context, proyekID domain.ID) ([]domain.Tugas, error)
}

func (r *Repository) CreateTugasNative(ctx context.Context, tugas domain.Tugas) error {
	if strings.TrimSpace(string(tugas.ID)) == "" ||
		strings.TrimSpace(string(tugas.RuangID)) == "" ||
		strings.TrimSpace(string(tugas.ProyekID)) == "" ||
		strings.TrimSpace(tugas.Title) == "" {
		return fmt.Errorf("tugas: %w", ErrInvalid)
	}

	var proyekRuangID string
	err := r.db.QueryRowContext(ctx, `SELECT ruang_id FROM proyek WHERE id = ?`, tugas.ProyekID).Scan(&proyekRuangID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("tugas: proyek tidak ditemukan: %w", ErrInvalid)
	}
	if err != nil {
		return fmt.Errorf("cek proyek tugas: %w", err)
	}
	if proyekRuangID != string(tugas.RuangID) {
		return fmt.Errorf("tugas: proyek tidak sesuai dengan ruang: %w", ErrInvalid)
	}

	var parentID any
	if tugas.ParentID != nil {
		value := strings.TrimSpace(string(*tugas.ParentID))
		if value == "" {
			return fmt.Errorf("tugas: parent tidak valid: %w", ErrInvalid)
		}
		var parentProyekID, parentRuangID string
		err := r.db.QueryRowContext(ctx, `SELECT proyek_id, ruang_id FROM tugas WHERE id = ?`, value).Scan(&parentProyekID, &parentRuangID)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("tugas: parent tidak ditemukan: %w", ErrInvalid)
		}
		if err != nil {
			return fmt.Errorf("cek parent tugas: %w", err)
		}
		if parentProyekID != string(tugas.ProyekID) || parentRuangID != string(tugas.RuangID) || value == string(tugas.ID) {
			return fmt.Errorf("tugas: parent tidak sesuai dengan proyek atau ruang: %w", ErrInvalid)
		}
		parentID = value
	}

	created, updated := timestamps(tugas.CreatedAt, tugas.UpdatedAt)
	_, err = r.db.ExecContext(ctx, `INSERT INTO tugas (id, ruang_id, proyek_id, parent_id, title, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, tugas.ID, tugas.RuangID, tugas.ProyekID, parentID, tugas.Title, tugas.Status, created, updated)
	if err != nil {
		return fmt.Errorf("buat tugas: %w", err)
	}
	return nil
}

func (r *Repository) GetTugasNative(ctx context.Context, id domain.ID) (domain.Tugas, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Tugas{}, fmt.Errorf("tugas: %w", ErrInvalid)
	}

	var tugas domain.Tugas
	var parentID sql.NullString
	var created, updated string
	err := r.db.QueryRowContext(ctx, `SELECT id, ruang_id, proyek_id, parent_id, title, status, created_at, updated_at FROM tugas WHERE id = ?`, id).Scan(&tugas.ID, &tugas.RuangID, &tugas.ProyekID, &parentID, &tugas.Title, &tugas.Status, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Tugas{}, ErrNotFound
	}
	if err != nil {
		return domain.Tugas{}, fmt.Errorf("ambil tugas: %w", err)
	}
	if parentID.Valid {
		value := domain.ID(parentID.String)
		tugas.ParentID = &value
	}
	var parseErr error
	if tugas.CreatedAt, parseErr = parseTime(created); parseErr != nil {
		return domain.Tugas{}, parseErr
	}
	if tugas.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
		return domain.Tugas{}, parseErr
	}
	return tugas, nil
}

func (r *Repository) ListTugasByProyek(ctx context.Context, proyekID domain.ID) ([]domain.Tugas, error) {
	if strings.TrimSpace(string(proyekID)) == "" {
		return nil, fmt.Errorf("tugas: proyek wajib diisi: %w", ErrInvalid)
	}

	rows, err := r.db.QueryContext(ctx, `SELECT id, ruang_id, proyek_id, parent_id, title, status, created_at, updated_at FROM tugas WHERE proyek_id = ? ORDER BY created_at, id`, proyekID)
	if err != nil {
		return nil, fmt.Errorf("daftar tugas: %w", err)
	}
	defer rows.Close()

	var result []domain.Tugas
	for rows.Next() {
		var tugas domain.Tugas
		var parentID sql.NullString
		var created, updated string
		if err := rows.Scan(&tugas.ID, &tugas.RuangID, &tugas.ProyekID, &parentID, &tugas.Title, &tugas.Status, &created, &updated); err != nil {
			return nil, fmt.Errorf("baca tugas: %w", err)
		}
		if parentID.Valid {
			value := domain.ID(parentID.String)
			tugas.ParentID = &value
		}
		var parseErr error
		if tugas.CreatedAt, parseErr = parseTime(created); parseErr != nil {
			return nil, parseErr
		}
		if tugas.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
			return nil, parseErr
		}
		result = append(result, tugas)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar tugas: %w", err)
	}
	return result, nil
}
