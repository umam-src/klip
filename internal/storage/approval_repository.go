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
	if strings.TrimSpace(string(approval.ID)) == "" || strings.TrimSpace(string(approval.RuangID)) == "" || strings.TrimSpace(string(approval.ProyekID)) == "" {
		return fmt.Errorf("persetujuan: identitas wajib diisi: %w", ErrInvalid)
	}
	if approval.Status == "" {
		approval.Status = domain.ApprovalPending
	}
	if approval.Status != domain.ApprovalPending {
		return fmt.Errorf("persetujuan: status awal harus pending: %w", ErrInvalid)
	}
	proyek, err := r.GetProyek(ctx, approval.ProyekID)
	if err != nil {
		return err
	}
	if proyek.RuangID != approval.RuangID {
		return fmt.Errorf("persetujuan: proyek tidak sesuai dengan ruang: %w", ErrInvalid)
	}
	if approval.TugasID != nil {
		tugas, err := r.GetTugasNative(ctx, *approval.TugasID)
		if err != nil {
			return err
		}
		if tugas.ProyekID != approval.ProyekID || tugas.RuangID != approval.RuangID {
			return fmt.Errorf("persetujuan: tugas tidak sesuai dengan proyek atau ruang: %w", ErrInvalid)
		}
	}
	created, updated := timestamps(approval.CreatedAt, approval.UpdatedAt)
	var tugasID any
	if approval.TugasID != nil {
		tugasID = string(*approval.TugasID)
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO persetujuan (id, ruang_id, proyek_id, tugas_id, status, reason, created_at, updated_at, decided_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, approval.ID, approval.RuangID, approval.ProyekID, tugasID, approval.Status, strings.TrimSpace(approval.Reason), created, updated, nil)
	if err != nil {
		return fmt.Errorf("buat persetujuan: %w", err)
	}
	return nil
}

func (r *Repository) GetApproval(ctx context.Context, id domain.ID) (domain.Approval, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Approval{}, fmt.Errorf("persetujuan: %w", ErrInvalid)
	}
	var a domain.Approval
	var tugasID, decidedAt sql.NullString
	var created, updated string
	err := r.db.QueryRowContext(ctx, `SELECT id, ruang_id, proyek_id, tugas_id, status, reason, created_at, updated_at, decided_at FROM persetujuan WHERE id = ?`, id).Scan(&a.ID, &a.RuangID, &a.ProyekID, &tugasID, &a.Status, &a.Reason, &created, &updated, &decidedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Approval{}, ErrNotFound
	}
	if err != nil {
		return domain.Approval{}, fmt.Errorf("ambil persetujuan: %w", err)
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

func (r *Repository) ListApprovalsByProyek(ctx context.Context, proyekID domain.ID) ([]domain.Approval, error) {
	if strings.TrimSpace(string(proyekID)) == "" {
		return nil, fmt.Errorf("persetujuan: proyek wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, ruang_id, proyek_id, tugas_id, status, reason, created_at, updated_at, decided_at FROM persetujuan WHERE proyek_id = ? ORDER BY created_at DESC, id DESC`, proyekID)
	if err != nil {
		return nil, fmt.Errorf("daftar persetujuan: %w", err)
	}
	defer rows.Close()
	return scanApprovalRows(rows)
}

func (r *Repository) ListApprovalsByTugas(ctx context.Context, tugasID domain.ID) ([]domain.Approval, error) {
	if strings.TrimSpace(string(tugasID)) == "" {
		return nil, fmt.Errorf("persetujuan: tugas wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, ruang_id, proyek_id, tugas_id, status, reason, created_at, updated_at, decided_at FROM persetujuan WHERE tugas_id = ? ORDER BY created_at DESC, id DESC`, tugasID)
	if err != nil {
		return nil, fmt.Errorf("daftar persetujuan tugas: %w", err)
	}
	defer rows.Close()
	return scanApprovalRows(rows)
}

func (r *Repository) DecideApproval(ctx context.Context, id domain.ID, status domain.ApprovalStatus, reason string) error {
	if status != domain.ApprovalApproved && status != domain.ApprovalRejected {
		return fmt.Errorf("persetujuan: keputusan tidak valid: %w", ErrInvalid)
	}
	current, err := r.GetApproval(ctx, id)
	if err != nil {
		return err
	}
	if current.Status != domain.ApprovalPending {
		return fmt.Errorf("persetujuan: keputusan hanya dapat dilakukan saat pending: %w", ErrInvalid)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err = r.db.ExecContext(ctx, `UPDATE persetujuan SET status = ?, reason = ?, updated_at = ?, decided_at = ? WHERE id = ? AND status = ?`, status, strings.TrimSpace(reason), now, now, id, domain.ApprovalPending)
	if err != nil {
		return fmt.Errorf("putuskan persetujuan: %w", err)
	}
	return nil
}

func (r *Repository) ApprovalGate(ctx context.Context, proyekID domain.ID, tugasID *domain.ID) (domain.ApprovalStatus, bool, error) {
	var status string
	var query string
	var args []any
	if tugasID != nil {
		query = `SELECT status FROM persetujuan WHERE tugas_id = ? ORDER BY created_at DESC, id DESC LIMIT 1`
		args = []any{*tugasID}
	} else {
		query = `SELECT status FROM persetujuan WHERE proyek_id = ? AND tugas_id IS NULL ORDER BY created_at DESC, id DESC LIMIT 1`
		args = []any{proyekID}
	}
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&status); errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	} else if err != nil {
		return "", false, fmt.Errorf("cek gerbang persetujuan: %w", err)
	}
	return domain.ApprovalStatus(status), true, nil
}

func scanApprovalRows(rows *sql.Rows) ([]domain.Approval, error) {
	var result []domain.Approval
	for rows.Next() {
		var a domain.Approval
		var tugasID, decidedAt sql.NullString
		var created, updated string
		if err := rows.Scan(&a.ID, &a.RuangID, &a.ProyekID, &tugasID, &a.Status, &a.Reason, &created, &updated, &decidedAt); err != nil {
			return nil, fmt.Errorf("baca persetujuan: %w", err)
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
		return nil, fmt.Errorf("baca daftar persetujuan: %w", err)
	}
	return result, nil
}
