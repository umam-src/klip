package storage

import (
	"context"
	"database/sql"
	"encoding/json"
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
		strings.TrimSpace(string(run.AgenID)) == "" || strings.TrimSpace(run.Program) == "" {
		return fmt.Errorf("run: %w", ErrInvalid)
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
		string(arguments), exitCode, run.Stdout, run.Stderr,
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
	if err == sql.ErrNoRows {
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
	if strings.TrimSpace(string(id)) == "" || strings.TrimSpace(string(status)) == "" {
		return fmt.Errorf("run: %w", ErrInvalid)
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
