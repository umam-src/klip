package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

type RunRepository interface {
	CreateRun(ctx context.Context, run domain.Run) error
	GetRun(ctx context.Context, id domain.ID) (domain.Run, error)
	ListRunsByPekerjaan(ctx context.Context, pekerjaanID domain.ID) ([]domain.Run, error)
	FinishRun(ctx context.Context, id domain.ID, status domain.Status, exitCode *int, stdout, stderr string, finishedAt time.Time) error
}

func (r *Repository) CreateRun(ctx context.Context, run domain.Run) error {
	if strings.TrimSpace(string(run.ID)) == "" || strings.TrimSpace(string(run.PekerjaanID)) == "" ||
		strings.TrimSpace(string(run.AgenID)) == "" || strings.TrimSpace(run.Program) == "" ||
		run.Status != domain.StatusRunning {
		return fmt.Errorf("run: %w", ErrInvalid)
	}
	var pekerjaanRuangID, agenRuangID string
	if err := r.db.QueryRowContext(ctx, `SELECT ruang_id FROM pekerjaan WHERE id = ?`, run.PekerjaanID).Scan(&pekerjaanRuangID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("run: pekerjaan tidak ditemukan: %w", ErrInvalid)
		}
		return fmt.Errorf("cek ruang pekerjaan run: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT ruang_id FROM agen WHERE id = ?`, run.AgenID).Scan(&agenRuangID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("run: agen tidak ditemukan: %w", ErrInvalid)
		}
		return fmt.Errorf("cek ruang agen run: %w", err)
	}
	if pekerjaanRuangID != agenRuangID {
		return fmt.Errorf("run: agen tidak sesuai dengan ruang pekerjaan: %w", ErrInvalid)
	}
	if run.TugasID != nil {
		value := strings.TrimSpace(string(*run.TugasID))
		if value == "" {
			return fmt.Errorf("run: tugas tidak valid: %w", ErrInvalid)
		}
		var tugasPekerjaanID string
		if err := r.db.QueryRowContext(ctx, `SELECT pekerjaan_id FROM tugas WHERE id = ?`, value).Scan(&tugasPekerjaanID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("run: tugas tidak ditemukan: %w", ErrInvalid)
			}
			return fmt.Errorf("cek tugas run: %w", err)
		}
		if tugasPekerjaanID != string(run.PekerjaanID) {
			return fmt.Errorf("run: tugas tidak sesuai dengan pekerjaan: %w", ErrInvalid)
		}
	}
	arguments, err := json.Marshal(run.Arguments)
	if err != nil {
		return fmt.Errorf("run: serialisasi argumen: %w", err)
	}
	started := run.StartedAt
	if started.IsZero() {
		started = time.Now().UTC()
	}
	var tugasID any
	if run.TugasID != nil {
		tugasID = string(*run.TugasID)
	}
	var exitCode any
	if run.ExitCode != nil {
		exitCode = *run.ExitCode
	}
	var finished any
	if run.FinishedAt != nil {
		finished = run.FinishedAt.UTC().Format(time.RFC3339Nano)
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO run (
			id, pekerjaan_id, tugas_id, agen_id, status, program, arguments,
			exit_code, stdout, stderr, started_at, finished_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		run.ID, run.PekerjaanID, tugasID, run.AgenID, run.Status, run.Program,
		run.Arguments, string(arguments), exitCode, run.Stdout, run.Stderr,
		started.UTC().Format(time.RFC3339Nano), finished,
	)
	if err != nil {
		return fmt.Errorf("buat run: %w", err)
	}
	return nil
}

func (r *Repository) GetRun(ctx context.Context, id domain.ID) (domain.Run, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Run{}, fmt.Errorf("run: %w", ErrInvalid)
	}
	var run domain.Run
	var tugasID, finished sql.NullString
	var exitCode sql.NullInt64
	var arguments, started string
	err := r.db.QueryRowContext(ctx, `
		SELECT id, pekerjaan_id, tugas_id, agen_id, status, program, arguments,
		       exit_code, stdout, stderr, started_at, finished_at
		FROM run WHERE id = ?`, id).Scan(
		&run.ID, &run.PekerjaanID, &tugasID, &run.AgenID, &run.Status,
		&run.Program, &arguments, &exitCode, &run.Stdout, &run.Stderr,
		&started, &finished,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Run{}, ErrNotFound
	}
	if err != nil {
		return domain.Run{}, fmt.Errorf("ambil run: %w", err)
	}
	if err := decodeRun(&run, tugasID, exitCode, arguments, started, finished); err != nil {
		return domain.Run{}, err
	}
	return run, nil
}

func (r *Repository) ListRunsByPekerjaan(ctx context.Context, pekerjaanID domain.ID) ([]domain.Run, error) {
	if strings.TrimSpace(string(pekerjaanID)) == "" {
		return nil, fmt.Errorf("run: pekerjaan wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, pekerjaan_id, tugas_id, agen_id, status, program, arguments,
		       exit_code, stdout, stderr, started_at, finished_at
		FROM run WHERE pekerjaan_id = ? ORDER BY started_at, id`, pekerjaanID)
	if err != nil {
		return nil, fmt.Errorf("daftar run: %w", err)
	}
	defer rows.Close()

	var result []domain.Run
	for rows.Next() {
		var run domain.Run
		var tugasID, finished sql.NullString
		var exitCode sql.NullInt64
		var arguments, started string
		if err := rows.Scan(
			&run.ID, &run.PekerjaanID, &tugasID, &run.AgenID, &run.Status,
			&run.Program, &arguments, &exitCode, &run.Stdout, &run.Stderr,
			&started, &finished,
		); err != nil {
			return nil, fmt.Errorf("baca run: %w", err)
		}
		if err := decodeRun(&run, tugasID, exitCode, arguments, started, finished); err != nil {
			return nil, err
		}
		result = append(result, run)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar run: %w", err)
	}
	return result, nil
}

func (r *Repository) FinishRun(ctx context.Context, id domain.ID, status domain.Status, exitCode *int, stdout, stderr string, finishedAt time.Time) error {
	if strings.TrimSpace(string(id)) == "" || !isTerminalStatus(status) {
		return fmt.Errorf("run: %w", ErrInvalid)
	}
	var current domain.Status
	if err := r.db.QueryRowContext(ctx, `SELECT status FROM run WHERE id = ?`, id).Scan(&current); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("baca status run: %w", err)
	}
	if current != domain.StatusRunning || !current.CanTransitionTo(status) {
		return fmt.Errorf("run: transisi status %q ke %q tidak diizinkan: %w", current, status, ErrInvalid)
	}
	if finishedAt.IsZero() {
		finishedAt = time.Now().UTC()
	}
	var exit any
	if exitCode != nil {
		exit = *exitCode
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE run
		SET status = ?, exit_code = ?, stdout = ?, stderr = ?, finished_at = ?
		WHERE id = ?`,
		status, exit, stdout, stderr, finishedAt.UTC().Format(time.RFC3339Nano), id,
	)
	if err != nil {
		return fmt.Errorf("selesaikan run: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrNotFound
	}
	return nil
}

func decodeRun(run *domain.Run, tugasID sql.NullString, exitCode sql.NullInt64, arguments, started string, finished sql.NullString) error {
	if tugasID.Valid {
		value := domain.ID(tugasID.String)
		run.TugasID = &value
	}
	if exitCode.Valid {
		value := int(exitCode.Int64)
		run.ExitCode = &value
	}
	if err := json.Unmarshal([]byte(arguments), &run.Arguments); err != nil {
		return fmt.Errorf("decode argumen run: %w", err)
	}
	var err error
	if run.StartedAt, err = parseTime(started); err != nil {
		return err
	}
	if finished.Valid {
		value, err := parseTime(finished.String)
		if err != nil {
			return err
		}
		run.FinishedAt = &value
	}
	return nil
}
