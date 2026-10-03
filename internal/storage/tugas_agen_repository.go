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

type TugasAgenAssignment struct {
	TugasID    domain.ID `json:"tugas_id"`
	AgenID     domain.ID `json:"agen_id"`
	AssignedAt time.Time `json:"assigned_at"`
}

func (r *Repository) ensureTugasAgenTable(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS tugas_agen (
    tugas_id TEXT PRIMARY KEY REFERENCES tugas(id) ON DELETE CASCADE,
    agen_id TEXT NOT NULL REFERENCES agen(id) ON DELETE RESTRICT,
    assigned_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_tugas_agen_agen ON tugas_agen(agen_id);
`)
	if err != nil {
		return fmt.Errorf("siapkan assignment tugas: %w", err)
	}
	return nil
}

func (r *Repository) AssignTugasToAgen(ctx context.Context, tugasID, agenID domain.ID, assignedAt time.Time) error {
	if err := r.ensureTugasAgenTable(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(string(tugasID)) == "" || strings.TrimSpace(string(agenID)) == "" {
		return fmt.Errorf("assignment tugas: %w", ErrInvalid)
	}

	var pekerjaanRuangID, agenRuangID string
	err := r.db.QueryRowContext(ctx, `
SELECT p.ruang_id, a.ruang_id
FROM tugas t
JOIN pekerjaan p ON p.id = t.pekerjaan_id
JOIN agen a ON a.id = ?
WHERE t.id = ?`, agenID, tugasID).Scan(&pekerjaanRuangID, &agenRuangID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("cek assignment tugas: %w", err)
	}
	if pekerjaanRuangID != agenRuangID {
		return fmt.Errorf("assignment tugas: agen dan pekerjaan harus berada di ruang yang sama: %w", ErrInvalid)
	}

	if assignedAt.IsZero() {
		assignedAt = time.Now().UTC()
	}
	_, err = r.db.ExecContext(ctx, `
INSERT INTO tugas_agen (tugas_id, agen_id, assigned_at) VALUES (?, ?, ?)
ON CONFLICT(tugas_id) DO UPDATE SET agen_id = excluded.agen_id, assigned_at = excluded.assigned_at`,
		tugasID, agenID, assignedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("simpan assignment tugas: %w", err)
	}
	return nil
}

func (r *Repository) UnassignTugas(ctx context.Context, tugasID domain.ID) error {
	if err := r.ensureTugasAgenTable(ctx); err != nil {
		return err
	}
	if strings.TrimSpace(string(tugasID)) == "" {
		return fmt.Errorf("assignment tugas: %w", ErrInvalid)
	}
	result, err := r.db.ExecContext(ctx, `DELETE FROM tugas_agen WHERE tugas_id = ?`, tugasID)
	if err != nil {
		return fmt.Errorf("hapus assignment tugas: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		var exists bool
		if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM tugas WHERE id = ?)`, tugasID).Scan(&exists); err != nil {
			return fmt.Errorf("cek tugas assignment: %w", err)
		}
		if !exists {
			return ErrNotFound
		}
	}
	return nil
}

func (r *Repository) GetTugasAssignment(ctx context.Context, tugasID domain.ID) (TugasAgenAssignment, error) {
	if err := r.ensureTugasAgenTable(ctx); err != nil {
		return TugasAgenAssignment{}, err
	}
	var assignment TugasAgenAssignment
	var assignedAt string
	err := r.db.QueryRowContext(ctx, `SELECT tugas_id, agen_id, assigned_at FROM tugas_agen WHERE tugas_id = ?`, tugasID).
		Scan(&assignment.TugasID, &assignment.AgenID, &assignedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return TugasAgenAssignment{}, ErrNotFound
	}
	if err != nil {
		return TugasAgenAssignment{}, fmt.Errorf("ambil assignment tugas: %w", err)
	}
	assignment.AssignedAt, err = parseTime(assignedAt)
	if err != nil {
		return TugasAgenAssignment{}, err
	}
	return assignment, nil
}

func (r *Repository) ListTugasByAgen(ctx context.Context, agenID domain.ID) ([]domain.Tugas, error) {
	if err := r.ensureTugasAgenTable(ctx); err != nil {
		return nil, err
	}
	if strings.TrimSpace(string(agenID)) == "" {
		return nil, fmt.Errorf("agen: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `
SELECT t.id, t.pekerjaan_id, t.parent_id, t.title, t.status, t.position, t.created_at, t.updated_at
FROM tugas t
JOIN tugas_agen ta ON ta.tugas_id = t.id
WHERE ta.agen_id = ?
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
		if err := rows.Scan(&tugas.ID, &tugas.PekerjaanID, &parentID, &tugas.Title, &tugas.Status, &tugas.Position, &created, &updated); err != nil {
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
