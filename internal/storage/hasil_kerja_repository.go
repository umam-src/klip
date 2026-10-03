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

// HasilKerjaRepository adalah batas penyimpanan native untuk hasil kerja.
type HasilKerjaRepository interface {
	CreateHasilKerja(ctx context.Context, hasil domain.HasilKerja) error
	GetHasilKerja(ctx context.Context, id domain.ID) (domain.HasilKerja, error)
	ListHasilKerjaByProyek(ctx context.Context, proyekID domain.ID) ([]domain.HasilKerja, error)
	ListHasilKerjaByEksekusi(ctx context.Context, executionID domain.ID) ([]domain.HasilKerja, error)
}

func (r *Repository) CreateHasilKerja(ctx context.Context, hasil domain.HasilKerja) error {
	if !hasil.Valid() {
		return fmt.Errorf("hasil kerja: %w", ErrInvalid)
	}

	var proyekRuangID string
	if err := r.db.QueryRowContext(ctx, `SELECT ruang_id FROM proyek WHERE id = ?`, hasil.ProyekID).Scan(&proyekRuangID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("hasil kerja: proyek tidak ditemukan: %w", ErrInvalid)
		}
		return fmt.Errorf("cek proyek hasil kerja: %w", err)
	}
	if proyekRuangID != string(hasil.RuangID) {
		return fmt.Errorf("hasil kerja: proyek tidak sesuai ruang: %w", ErrInvalid)
	}

	if hasil.TugasID != nil {
		var tugasRuangID, tugasProyekID string
		if err := r.db.QueryRowContext(ctx, `SELECT ruang_id, proyek_id FROM tugas WHERE id = ?`, *hasil.TugasID).Scan(&tugasRuangID, &tugasProyekID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("hasil kerja: tugas tidak ditemukan: %w", ErrInvalid)
			}
			return fmt.Errorf("cek tugas hasil kerja: %w", err)
		}
		if tugasRuangID != string(hasil.RuangID) || tugasProyekID != string(hasil.ProyekID) {
			return fmt.Errorf("hasil kerja: tugas tidak sesuai proyek atau ruang: %w", ErrInvalid)
		}
	}

	if hasil.ExecutionID != nil {
		var executionRuangID, executionProyekID string
		if err := r.db.QueryRowContext(ctx, `SELECT ruang_id, proyek_id FROM eksekusi WHERE id = ?`, *hasil.ExecutionID).Scan(&executionRuangID, &executionProyekID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("hasil kerja: eksekusi tidak ditemukan: %w", ErrInvalid)
			}
			return fmt.Errorf("cek eksekusi hasil kerja: %w", err)
		}
		if executionRuangID != string(hasil.RuangID) || executionProyekID != string(hasil.ProyekID) {
			return fmt.Errorf("hasil kerja: eksekusi tidak sesuai proyek atau ruang: %w", ErrInvalid)
		}
	}

	created := hasil.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO hasil (id, ruang_id, execution_id, proyek_id, tugas_id, kind, name, path, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		hasil.ID, hasil.RuangID, optionalID(hasil.ExecutionID), hasil.ProyekID, optionalID(hasil.TugasID), hasil.Kind, hasil.Name, hasil.Path, created.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("buat hasil kerja: %w", err)
	}
	return nil
}

func (r *Repository) GetHasilKerja(ctx context.Context, id domain.ID) (domain.HasilKerja, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.HasilKerja{}, fmt.Errorf("hasil kerja: %w", ErrInvalid)
	}
	var hasil domain.HasilKerja
	var executionID, tugasID sql.NullString
	var created string
	err := r.db.QueryRowContext(ctx, `SELECT id, ruang_id, execution_id, proyek_id, tugas_id, kind, name, path, created_at FROM hasil WHERE id = ?`, id).Scan(
		&hasil.ID, &hasil.RuangID, &executionID, &hasil.ProyekID, &tugasID, &hasil.Kind, &hasil.Name, &hasil.Path, &created,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.HasilKerja{}, ErrNotFound
	}
	if err != nil {
		return domain.HasilKerja{}, fmt.Errorf("ambil hasil kerja: %w", err)
	}
	if executionID.Valid {
		value := domain.ID(executionID.String)
		hasil.ExecutionID = &value
	}
	if tugasID.Valid {
		value := domain.ID(tugasID.String)
		hasil.TugasID = &value
	}
	parsed, err := parseTime(created)
	if err != nil {
		return domain.HasilKerja{}, err
	}
	hasil.CreatedAt = parsed
	return hasil, nil
}

func (r *Repository) ListHasilKerjaByProyek(ctx context.Context, proyekID domain.ID) ([]domain.HasilKerja, error) {
	if strings.TrimSpace(string(proyekID)) == "" {
		return nil, fmt.Errorf("hasil kerja: proyek wajib diisi: %w", ErrInvalid)
	}
	return r.listHasilKerja(ctx, `WHERE proyek_id = ?`, proyekID)
}

func (r *Repository) ListHasilKerjaByEksekusi(ctx context.Context, executionID domain.ID) ([]domain.HasilKerja, error) {
	if strings.TrimSpace(string(executionID)) == "" {
		return nil, fmt.Errorf("hasil kerja: eksekusi wajib diisi: %w", ErrInvalid)
	}
	return r.listHasilKerja(ctx, `WHERE execution_id = ?`, executionID)
}

func (r *Repository) listHasilKerja(ctx context.Context, where string, arg any) ([]domain.HasilKerja, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, ruang_id, execution_id, proyek_id, tugas_id, kind, name, path, created_at FROM hasil `+where+` ORDER BY created_at, id`, arg)
	if err != nil {
		return nil, fmt.Errorf("daftar hasil kerja: %w", err)
	}
	defer rows.Close()

	var result []domain.HasilKerja
	for rows.Next() {
		var hasil domain.HasilKerja
		var executionID, tugasID sql.NullString
		var created string
		if err := rows.Scan(&hasil.ID, &hasil.RuangID, &executionID, &hasil.ProyekID, &tugasID, &hasil.Kind, &hasil.Name, &hasil.Path, &created); err != nil {
			return nil, fmt.Errorf("baca hasil kerja: %w", err)
		}
		if executionID.Valid {
			value := domain.ID(executionID.String)
			hasil.ExecutionID = &value
		}
		if tugasID.Valid {
			value := domain.ID(tugasID.String)
			hasil.TugasID = &value
		}
		parsed, err := parseTime(created)
		if err != nil {
			return nil, err
		}
		hasil.CreatedAt = parsed
		result = append(result, hasil)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar hasil kerja: %w", err)
	}
	return result, nil
}
