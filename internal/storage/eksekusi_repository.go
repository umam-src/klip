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

// EksekusiRepository adalah batas penyimpanan native untuk riwayat eksekusi.
type EksekusiRepository interface {
	CreateEksekusi(ctx context.Context, eksekusi domain.Eksekusi) error
	GetEksekusi(ctx context.Context, id domain.ID) (domain.Eksekusi, error)
	ListEksekusiByTugas(ctx context.Context, tugasID domain.ID) ([]domain.Eksekusi, error)
	ListEksekusiByProyek(ctx context.Context, proyekID domain.ID) ([]domain.Eksekusi, error)
	UpdateEksekusi(ctx context.Context, id domain.ID, status domain.Status, exitCode *int, stdout, stderr string, finishedAt time.Time) error
}

func (r *Repository) CreateEksekusi(ctx context.Context, eksekusi domain.Eksekusi) error {
	if err := validateEksekusi(eksekusi); err != nil {
		return fmt.Errorf("eksekusi: %w", err)
	}
	var proyekRuangID, agenRuangID string
	if err := r.db.QueryRowContext(ctx, `SELECT ruang_id FROM proyek WHERE id = ?`, eksekusi.ProyekID).Scan(&proyekRuangID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("eksekusi: proyek tidak ditemukan: %w", ErrInvalid)
		}
		return fmt.Errorf("cek proyek eksekusi: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT ruang_id FROM agen WHERE id = ?`, eksekusi.AgenID).Scan(&agenRuangID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("eksekusi: agen tidak ditemukan: %w", ErrInvalid)
		}
		return fmt.Errorf("cek agen eksekusi: %w", err)
	}
	if proyekRuangID != string(eksekusi.RuangID) || agenRuangID != string(eksekusi.RuangID) {
		return fmt.Errorf("eksekusi: konteks ruang tidak konsisten: %w", ErrInvalid)
	}
	if eksekusi.TugasID != nil {
		var tugasRuangID, tugasProyekID string
		if err := r.db.QueryRowContext(ctx, `SELECT ruang_id, proyek_id FROM tugas WHERE id = ?`, *eksekusi.TugasID).Scan(&tugasRuangID, &tugasProyekID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("eksekusi: tugas tidak ditemukan: %w", ErrInvalid)
			}
			return fmt.Errorf("cek tugas eksekusi: %w", err)
		}
		if tugasRuangID != string(eksekusi.RuangID) || tugasProyekID != string(eksekusi.ProyekID) {
			return fmt.Errorf("eksekusi: tugas tidak sesuai proyek atau ruang: %w", ErrInvalid)
		}
	}
	if eksekusi.StartedAt.IsZero() {
		return fmt.Errorf("eksekusi: waktu mulai wajib diisi: %w", ErrInvalid)
	}
	var finished any
	if eksekusi.FinishedAt != nil {
		finished = eksekusi.FinishedAt.UTC().Format(timeFormat)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO eksekusi (id, ruang_id, proyek_id, tugas_id, agen_id, status, program, arguments, exit_code, stdout, stderr, started_at, finished_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, eksekusi.ID, eksekusi.RuangID, eksekusi.ProyekID, optionalID(eksekusi.TugasID), eksekusi.AgenID, eksekusi.Status, eksekusi.Program, encodeArguments(eksekusi.Arguments), optionalExitCode(eksekusi.ExitCode), eksekusi.Stdout, eksekusi.Stderr, eksekusi.StartedAt.UTC().Format(timeFormat), finished)
	if err != nil {
		return fmt.Errorf("buat eksekusi: %w", err)
	}
	return nil
}

func (r *Repository) UpdateEksekusi(ctx context.Context, id domain.ID, status domain.Status, exitCode *int, stdout, stderr string, finishedAt time.Time) error {
	if strings.TrimSpace(string(id)) == "" {
		return fmt.Errorf("eksekusi: id wajib diisi: %w", ErrInvalid)
	}
	if !isKnownStatus(status) {
		return fmt.Errorf("eksekusi: status tidak valid: %w", ErrInvalid)
	}
	if finishedAt.IsZero() {
		return fmt.Errorf("eksekusi: waktu selesai wajib diisi: %w", ErrInvalid)
	}
	result, err := r.db.ExecContext(ctx, `UPDATE eksekusi SET status = ?, exit_code = ?, stdout = ?, stderr = ?, finished_at = ? WHERE id = ?`, status, optionalExitCode(exitCode), stdout, stderr, finishedAt.UTC().Format(timeFormat), id)
	if err != nil {
		return fmt.Errorf("perbarui eksekusi: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("cek pembaruan eksekusi: %w", err)
	} else if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) GetEksekusi(ctx context.Context, id domain.ID) (domain.Eksekusi, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Eksekusi{}, fmt.Errorf("eksekusi: %w", ErrInvalid)
	}
	return r.getEksekusi(ctx, `WHERE e.id = ?`, id)
}

func (r *Repository) ListEksekusiByTugas(ctx context.Context, tugasID domain.ID) ([]domain.Eksekusi, error) {
	if strings.TrimSpace(string(tugasID)) == "" {
		return nil, fmt.Errorf("eksekusi: tugas wajib diisi: %w", ErrInvalid)
	}
	return r.listEksekusi(ctx, `WHERE e.tugas_id = ?`, tugasID)
}

func (r *Repository) ListEksekusiByProyek(ctx context.Context, proyekID domain.ID) ([]domain.Eksekusi, error) {
	if strings.TrimSpace(string(proyekID)) == "" {
		return nil, fmt.Errorf("eksekusi: proyek wajib diisi: %w", ErrInvalid)
	}
	return r.listEksekusi(ctx, `WHERE e.proyek_id = ?`, proyekID)
}

const timeFormat = time.RFC3339Nano

func (r *Repository) getEksekusi(ctx context.Context, where string, arg any) (domain.Eksekusi, error) {
	row := r.db.QueryRowContext(ctx, `SELECT e.id, e.ruang_id, e.proyek_id, e.tugas_id, e.agen_id, e.status, e.program, e.arguments, e.exit_code, e.stdout, e.stderr, e.started_at, e.finished_at FROM eksekusi e `+where, arg)
	var e domain.Eksekusi
	var tugasID, finished sql.NullString
	var exit sql.NullInt64
	var args, started string
	if err := row.Scan(&e.ID, &e.RuangID, &e.ProyekID, &tugasID, &e.AgenID, &e.Status, &e.Program, &args, &exit, &e.Stdout, &e.Stderr, &started, &finished); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Eksekusi{}, ErrNotFound
		}
		return domain.Eksekusi{}, fmt.Errorf("ambil eksekusi: %w", err)
	}
	parsed, err := parseTime(started)
	if err != nil {
		return domain.Eksekusi{}, err
	}
	e.StartedAt = parsed
	if tugasID.Valid {
		v := domain.ID(tugasID.String)
		e.TugasID = &v
	}
	if exit.Valid {
		v := int(exit.Int64)
		e.ExitCode = &v
	}
	if finished.Valid {
		v, err := parseTime(finished.String)
		if err != nil {
			return domain.Eksekusi{}, err
		}
		e.FinishedAt = &v
	}
	e.Arguments = decodeArguments(args)
	return e, nil
}

func (r *Repository) listEksekusi(ctx context.Context, where string, arg any) ([]domain.Eksekusi, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT e.id, e.ruang_id, e.proyek_id, e.tugas_id, e.agen_id, e.status, e.program, e.arguments, e.exit_code, e.stdout, e.stderr, e.started_at, e.finished_at FROM eksekusi e `+where+` ORDER BY e.started_at, e.id`, arg)
	if err != nil {
		return nil, fmt.Errorf("daftar eksekusi: %w", err)
	}
	defer rows.Close()
	var result []domain.Eksekusi
	for rows.Next() {
		var e domain.Eksekusi
		var tugasID, finished sql.NullString
		var exit sql.NullInt64
		var args, started string
		if err := rows.Scan(&e.ID, &e.RuangID, &e.ProyekID, &tugasID, &e.AgenID, &e.Status, &args, &exit, &e.Stdout, &e.Stderr, &started, &finished); err != nil {
			return nil, fmt.Errorf("baca eksekusi: %w", err)
		}
		var err error
		e.StartedAt, err = parseTime(started)
		if err != nil {
			return nil, err
		}
		if tugasID.Valid {
			v := domain.ID(tugasID.String)
			e.TugasID = &v
		}
		if exit.Valid {
			v := int(exit.Int64)
			e.ExitCode = &v
		}
		if finished.Valid {
			v, err := parseTime(finished.String)
			if err != nil {
				return nil, err
			}
			e.FinishedAt = &v
		}
		e.Arguments = decodeArguments(args)
		result = append(result, e)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar eksekusi: %w", err)
	}
	return result, nil
}

func validateEksekusi(e domain.Eksekusi) error {
	if strings.TrimSpace(string(e.ID)) == "" || strings.TrimSpace(string(e.RuangID)) == "" || strings.TrimSpace(string(e.ProyekID)) == "" || strings.TrimSpace(string(e.AgenID)) == "" || strings.TrimSpace(e.Program) == "" {
		return ErrInvalid
	}
	switch e.Status {
	case domain.StatusDraft, domain.StatusReady, domain.StatusRunning, domain.StatusWaiting, domain.StatusBlocked, domain.StatusCompleted, domain.StatusFailed, domain.StatusCancelled:
	default:
		return ErrInvalid
	}
	return nil
}

func optionalID(id *domain.ID) any {
	if id == nil {
		return nil
	}
	return string(*id)
}

func optionalExitCode(code *int) any {
	if code == nil {
		return nil
	}
	return *code
}

func encodeArguments(values []string) string { return strings.Join(values, "\x00") }

func decodeArguments(value string) []string {
	if value == "" {
		return nil
	}
	return strings.Split(value, "\x00")
}
