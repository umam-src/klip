package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/umam-src/klip/internal/domain"
)

const komentarColumns = `id, ruang_id, proyek_id, tugas_id, parent_id, body, created_at, updated_at`

func (r *Repository) CreateKomentar(ctx context.Context, komentar domain.Komentar) error {
	if strings.TrimSpace(string(komentar.ID)) == "" || strings.TrimSpace(komentar.Body) == "" {
		return fmt.Errorf("komentar: id dan isi wajib diisi: %w", ErrInvalid)
	}
	if strings.TrimSpace(string(komentar.ProyekID)) == "" {
		return fmt.Errorf("komentar: proyek wajib diisi: %w", ErrInvalid)
	}
	var proyekRuangID string
	if err := r.db.QueryRowContext(ctx, `SELECT ruang_id FROM proyek WHERE id = ?`, komentar.ProyekID).Scan(&proyekRuangID); errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	} else if err != nil {
		return fmt.Errorf("cek proyek komentar: %w", err)
	}
	if proyekRuangID != string(komentar.RuangID) {
		return fmt.Errorf("komentar: ruang kerja tidak sesuai dengan proyek: %w", ErrInvalid)
	}
	if komentar.TugasID != nil {
		var tugasProyekID string
		if err := r.db.QueryRowContext(ctx, `SELECT proyek_id FROM tugas WHERE id = ?`, *komentar.TugasID).Scan(&tugasProyekID); errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		} else if err != nil {
			return fmt.Errorf("cek tugas komentar: %w", err)
		}
		if tugasProyekID != string(komentar.ProyekID) {
			return fmt.Errorf("komentar: tugas tidak sesuai dengan proyek: %w", ErrInvalid)
		}
	}
	if komentar.ParentID != nil {
		var parentProyekID string
		var parentTugasID sql.NullString
		if err := r.db.QueryRowContext(ctx, `SELECT proyek_id, tugas_id FROM komentar WHERE id = ?`, *komentar.ParentID).Scan(&parentProyekID, &parentTugasID); errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("komentar: induk tidak ditemukan: %w", ErrInvalid)
		} else if err != nil {
			return fmt.Errorf("cek induk komentar: %w", err)
		}
		if parentProyekID != string(komentar.ProyekID) || !sameOptionalKomentarID(parentTugasID, komentar.TugasID) {
			return fmt.Errorf("komentar: induk tidak sesuai konteks: %w", ErrInvalid)
		}
	}
	created, updated := timestamps(komentar.CreatedAt, komentar.UpdatedAt)
	var tugasID, parentID any
	if komentar.TugasID != nil {
		tugasID = string(*komentar.TugasID)
	}
	if komentar.ParentID != nil {
		parentID = string(*komentar.ParentID)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO komentar (`+komentarColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, komentar.ID, komentar.RuangID, komentar.ProyekID, tugasID, parentID, strings.TrimSpace(komentar.Body), created, updated)
	if err != nil {
		return fmt.Errorf("buat komentar: %w", err)
	}
	return nil
}

func (r *Repository) GetKomentar(ctx context.Context, id domain.ID) (domain.Komentar, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Komentar{}, fmt.Errorf("komentar: %w", ErrInvalid)
	}
	var k domain.Komentar
	var tugasID, parentID sql.NullString
	var created, updated string
	err := r.db.QueryRowContext(ctx, `SELECT `+komentarColumns+` FROM komentar WHERE id = ?`, id).Scan(&k.ID, &k.RuangID, &k.ProyekID, &tugasID, &parentID, &k.Body, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Komentar{}, ErrNotFound
	}
	if err != nil {
		return domain.Komentar{}, fmt.Errorf("ambil komentar: %w", err)
	}
	if tugasID.Valid {
		v := domain.ID(tugasID.String)
		k.TugasID = &v
	}
	if parentID.Valid {
		v := domain.ID(parentID.String)
		k.ParentID = &v
	}
	var parseErr error
	if k.CreatedAt, parseErr = parseTime(created); parseErr != nil {
		return domain.Komentar{}, parseErr
	}
	if k.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
		return domain.Komentar{}, parseErr
	}
	return k, nil
}

// ListKomentarByProyek mengembalikan komentar tingkat Proyek, yaitu komentar
// yang tidak terikat pada Tugas tertentu.
func (r *Repository) ListKomentarByProyek(ctx context.Context, proyekID domain.ID) ([]domain.Komentar, error) {
	if strings.TrimSpace(string(proyekID)) == "" {
		return nil, fmt.Errorf("komentar: proyek wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+komentarColumns+` FROM komentar WHERE proyek_id = ? AND tugas_id IS NULL ORDER BY created_at, id`, proyekID)
	if err != nil {
		return nil, fmt.Errorf("daftar komentar: %w", err)
	}
	defer rows.Close()
	return scanKomentarRows(rows)
}

func (r *Repository) ListKomentarByTugas(ctx context.Context, tugasID domain.ID) ([]domain.Komentar, error) {
	if strings.TrimSpace(string(tugasID)) == "" {
		return nil, fmt.Errorf("komentar: tugas wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT `+komentarColumns+` FROM komentar WHERE tugas_id = ? ORDER BY created_at, id`, tugasID)
	if err != nil {
		return nil, fmt.Errorf("daftar komentar tugas: %w", err)
	}
	defer rows.Close()
	return scanKomentarRows(rows)
}

func scanKomentarRows(rows *sql.Rows) ([]domain.Komentar, error) {
	var result []domain.Komentar
	for rows.Next() {
		var k domain.Komentar
		var tugasID, parentID sql.NullString
		var created, updated string
		if err := rows.Scan(&k.ID, &k.RuangID, &k.ProyekID, &tugasID, &parentID, &k.Body, &created, &updated); err != nil {
			return nil, fmt.Errorf("baca komentar: %w", err)
		}
		if tugasID.Valid {
			v := domain.ID(tugasID.String)
			k.TugasID = &v
		}
		if parentID.Valid {
			v := domain.ID(parentID.String)
			k.ParentID = &v
		}
		var err error
		if k.CreatedAt, err = parseTime(created); err != nil {
			return nil, err
		}
		if k.UpdatedAt, err = parseTime(updated); err != nil {
			return nil, err
		}
		result = append(result, k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar komentar: %w", err)
	}
	return result, nil
}

func sameOptionalKomentarID(value sql.NullString, expected *domain.ID) bool {
	if !value.Valid {
		return expected == nil
	}
	return expected != nil && value.String == string(*expected)
}
