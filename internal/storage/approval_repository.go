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

func (r *Repository) CreateApproval(ctx context.Context, approval domain.Approval) error {
	if strings.TrimSpace(string(approval.ID)) == "" || strings.TrimSpace(string(approval.PekerjaanID)) == "" {
		return fmt.Errorf("approval: id dan pekerjaan wajib diisi: %w", ErrInvalid)
	}
	if approval.Status == "" {
		approval.Status = domain.ApprovalPending
	}
	if approval.Status != domain.ApprovalPending {
		return fmt.Errorf("approval: status awal harus pending: %w", ErrInvalid)
	}
	if _, err := r.GetPekerjaan(ctx, approval.PekerjaanID); err != nil {
		return err
	}
	if approval.TugasID != nil {
		tugas, err := r.GetTugas(ctx, *approval.TugasID)
		if err != nil {
			return err
		}
		if tugas.PekerjaanID != approval.PekerjaanID {
			return fmt.Errorf("approval: tugas tidak sesuai dengan pekerjaan: %w", ErrInvalid)
		}
	}
	created, updated := timestamps(approval.CreatedAt, approval.UpdatedAt)
	var tugasID any
	if approval.TugasID != nil {
		tugasID = string(*approval.TugasID)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO approval (id, pekerjaan_id, tugas_id, status, reason, created_at, updated_at, decided_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, approval.ID, approval.PekerjaanID, tugasID, approval.Status, strings.TrimSpace(approval.Reason), created, updated, nil)
	if err != nil {
		return fmt.Errorf("buat approval: %w", err)
	}
	return nil
}

func (r *Repository) GetApproval(ctx context.Context, id domain.ID) (domain.Approval, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Approval{}, fmt.Errorf("approval: %w", ErrInvalid)
	}
	var a domain.Approval
	var tugasID, decidedAt sql.NullString
	var created, updated string
	err := r.db.QueryRowContext(ctx, `SELECT id, pekerjaan_id, tugas_id, status, reason, created_at, updated_at, decided_at FROM approval WHERE id = ?`, id).Scan(&a.ID, &a.PekerjaanID, &tugasID, &a.Status, &a.Reason, &created, &updated, &decidedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Approval{}, ErrNotFound
	}
	if err != nil {
		return domain.Approval{}, fmt.Errorf("ambil approval: %w", err)
	}
	if tugasID.Valid {
		v := domain.ID(tugasID.String)
		a.TugasID = &v
	}
	var parseErr error
	if a.CreatedAt, parseErr = parseTime(created); parseErr != nil {
		return domain.Approval{}, parseErr
	}
	if a.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
		return domain.Approval{}, parseErr
	}
	if decidedAt.Valid {
		v, err := parseTime(decidedAt.String)
		if err != nil {
			return domain.Approval{}, err
		}
		a.DecidedAt = &v
	}
	return a, nil
}

func (r *Repository) ListApprovalsByPekerjaan(ctx context.Context, pekerjaanID domain.ID) ([]domain.Approval, error) {
	if strings.TrimSpace(string(pekerjaanID)) == "" {
		return nil, fmt.Errorf("approval: pekerjaan wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, pekerjaan_id, tugas_id, status, reason, created_at, updated_at, decided_at FROM approval WHERE pekerjaan_id = ? ORDER BY created_at DESC, id DESC`, pekerjaanID)
	if err != nil {
		return nil, fmt.Errorf("daftar approval: %w", err)
	}
	defer rows.Close()
	return scanApprovalRows(rows)
}

func (r *Repository) ListApprovalsByTugas(ctx context.Context, tugasID domain.ID) ([]domain.Approval, error) {
	if strings.TrimSpace(string(tugasID)) == "" {
		return nil, fmt.Errorf("approval: tugas wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, pekerjaan_id, tugas_id, status, reason, created_at, updated_at, decided_at FROM approval WHERE tugas_id = ? ORDER BY created_at DESC, id DESC`, tugasID)
	if err != nil {
		return nil, fmt.Errorf("daftar approval tugas: %w", err)
	}
	defer rows.Close()
	return scanApprovalRows(rows)
}

func (r *Repository) DecideApproval(ctx context.Context, id domain.ID, status domain.ApprovalStatus, reason string) error {
	if status != domain.ApprovalApproved && status != domain.ApprovalRejected {
		return fmt.Errorf("approval: keputusan tidak valid: %w", ErrInvalid)
	}
	current, err := r.GetApproval(ctx, id)
	if err != nil {
		return err
	}
	if current.Status != domain.ApprovalPending {
		return fmt.Errorf("approval: keputusan hanya dapat dilakukan saat pending: %w", ErrInvalid)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = r.db.ExecContext(ctx, `UPDATE approval SET status = ?, reason = ?, updated_at = ?, decided_at = ? WHERE id = ? AND status = ?`, status, strings.TrimSpace(reason), now, now, id, domain.ApprovalPending)
	if err != nil {
		return fmt.Errorf("putuskan approval: %w", err)
	}
	return nil
}

func (r *Repository) ApprovalGate(ctx context.Context, pekerjaanID domain.ID, tugasID *domain.ID) (domain.ApprovalStatus, bool, error) {
	var status string
	var query string
	var args []any
	if tugasID != nil {
		query = `SELECT status FROM approval WHERE tugas_id = ? ORDER BY created_at DESC, id DESC LIMIT 1`
		args = []any{*tugasID}
	} else {
		query = `SELECT status FROM approval WHERE pekerjaan_id = ? AND tugas_id IS NULL ORDER BY created_at DESC, id DESC LIMIT 1`
		args = []any{pekerjaanID}
	}
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&status)
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, fmt.Errorf("cek approval gate: %w", err)
	}
	return domain.ApprovalStatus(status), true, nil
}

func scanApprovalRows(rows *sql.Rows) ([]domain.Approval, error) {
	var result []domain.Approval
	for rows.Next() {
		var a domain.Approval
		var tugasID, decidedAt sql.NullString
		var created, updated string
		if err := rows.Scan(&a.ID, &a.PekerjaanID, &tugasID, &a.Status, &a.Reason, &created, &updated, &decidedAt); err != nil {
			return nil, fmt.Errorf("baca approval: %w", err)
		}
		if tugasID.Valid {
			v := domain.ID(tugasID.String)
			a.TugasID = &v
		}
		var err error
		if a.CreatedAt, err = parseTime(created); err != nil {
			return nil, err
		}
		if a.UpdatedAt, err = parseTime(updated); err != nil {
			return nil, err
		}
		if decidedAt.Valid {
			v, err := parseTime(decidedAt.String)
			if err != nil {
				return nil, err
			}
			a.DecidedAt = &v
		}
		result = append(result, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar approval: %w", err)
	}
	return result, nil
}
