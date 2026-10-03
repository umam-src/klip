package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

// TugasAgenAssignment adalah tampilan ringkas Penugasan untuk satu Tugas.
// AssignedAt berasal dari waktu pembuatan Penugasan.
type TugasAgenAssignment struct {
	TugasID    domain.ID `json:"tugas_id"`
	AgenID     domain.ID `json:"agen_id"`
	AssignedAt time.Time `json:"assigned_at"`
}

func newPenugasanID() domain.ID {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return domain.ID(fmt.Sprintf("penugasan-%d", time.Now().UnixNano()))
	}
	return domain.ID("penugasan-" + hex.EncodeToString(data[:]))
}

// AssignTugasToAgen menetapkan satu Agen pelaksana untuk Tugas dengan mengganti
// Penugasan yang ada. Tugas dan Agen harus berada di Ruang Kerja yang sama.
func (r *Repository) AssignTugasToAgen(ctx context.Context, tugasID, agenID domain.ID, assignedAt time.Time) error {
	if strings.TrimSpace(string(tugasID)) == "" || strings.TrimSpace(string(agenID)) == "" {
		return fmt.Errorf("penugasan tugas: %w", ErrInvalid)
	}

	var tugasRuangID, agenRuangID string
	err := r.db.QueryRowContext(ctx, `
SELECT t.ruang_id, a.ruang_id
FROM tugas t
JOIN agen a ON a.id = ?
WHERE t.id = ?`, agenID, tugasID).Scan(&tugasRuangID, &agenRuangID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("cek penugasan tugas: %w", err)
	}
	if tugasRuangID != agenRuangID {
		return fmt.Errorf("penugasan tugas: agen dan tugas harus berada di ruang kerja yang sama: %w", ErrInvalid)
	}

	if assignedAt.IsZero() {
		assignedAt = time.Now().UTC()
	}
	stamp := assignedAt.UTC().Format(time.RFC3339Nano)

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("mulai penugasan tugas: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM penugasan WHERE tugas_id = ?`, tugasID); err != nil {
		return fmt.Errorf("ganti penugasan tugas: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO penugasan (id, tugas_id, agen_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, newPenugasanID(), tugasID, agenID, stamp, stamp); err != nil {
		return fmt.Errorf("simpan penugasan tugas: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("simpan penugasan tugas: %w", err)
	}
	return nil
}

// UnassignTugas menghapus seluruh Penugasan pada Tugas.
func (r *Repository) UnassignTugas(ctx context.Context, tugasID domain.ID) error {
	if strings.TrimSpace(string(tugasID)) == "" {
		return fmt.Errorf("penugasan tugas: %w", ErrInvalid)
	}
	result, err := r.db.ExecContext(ctx, `DELETE FROM penugasan WHERE tugas_id = ?`, tugasID)
	if err != nil {
		return fmt.Errorf("hapus penugasan tugas: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		var exists bool
		if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM tugas WHERE id = ?)`, tugasID).Scan(&exists); err != nil {
			return fmt.Errorf("cek tugas penugasan: %w", err)
		}
		if !exists {
			return ErrNotFound
		}
	}
	return nil
}

// GetTugasAssignment mengembalikan Penugasan paling awal pada Tugas.
func (r *Repository) GetTugasAssignment(ctx context.Context, tugasID domain.ID) (TugasAgenAssignment, error) {
	var assignment TugasAgenAssignment
	var assignedAt string
	err := r.db.QueryRowContext(ctx, `SELECT tugas_id, agen_id, created_at FROM penugasan WHERE tugas_id = ? ORDER BY created_at, id LIMIT 1`, tugasID).
		Scan(&assignment.TugasID, &assignment.AgenID, &assignedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return TugasAgenAssignment{}, ErrNotFound
	}
	if err != nil {
		return TugasAgenAssignment{}, fmt.Errorf("ambil penugasan tugas: %w", err)
	}
	assignment.AssignedAt, err = parseTime(assignedAt)
	if err != nil {
		return TugasAgenAssignment{}, err
	}
	return assignment, nil
}

// ListTugasByAgen mengembalikan Tugas yang ditugaskan kepada Agen.
func (r *Repository) ListTugasByAgen(ctx context.Context, agenID domain.ID) ([]domain.Tugas, error) {
	if strings.TrimSpace(string(agenID)) == "" {
		return nil, fmt.Errorf("agen: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT t.id, t.ruang_id, t.proyek_id, t.parent_id, t.title, t.status, t.created_at, t.updated_at
FROM tugas t
JOIN penugasan pn ON pn.tugas_id = t.id
WHERE pn.agen_id = ?
ORDER BY t.created_at, t.id`, agenID)
	if err != nil {
		return nil, fmt.Errorf("daftar tugas agen: %w", err)
	}
	defer rows.Close()

	var result []domain.Tugas
	for rows.Next() {
		var tugas domain.Tugas
		var parentID sql.NullString
		var created, updated string
		if err := rows.Scan(&tugas.ID, &tugas.RuangID, &tugas.ProyekID, &parentID, &tugas.Title, &tugas.Status, &created, &updated); err != nil {
			return nil, fmt.Errorf("baca tugas agen: %w", err)
		}
		if parentID.Valid {
			value := domain.ID(parentID.String)
			tugas.ParentID = &value
		}
		if tugas.CreatedAt, err = parseTime(created); err != nil {
			return nil, err
		}
		if tugas.UpdatedAt, err = parseTime(updated); err != nil {
			return nil, err
		}
		result = append(result, tugas)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar tugas agen: %w", err)
	}
	return result, nil
}
