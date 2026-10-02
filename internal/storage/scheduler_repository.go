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

type SchedulerRepository interface {
	CreateSchedule(ctx context.Context, schedule domain.Schedule) error
	GetSchedule(ctx context.Context, id domain.ID) (domain.Schedule, error)
	ListSchedules(ctx context.Context) ([]domain.Schedule, error)
	UpdateScheduleRun(ctx context.Context, id domain.ID, nextRunAt time.Time) error
	CreateScheduleRun(ctx context.Context, run domain.ScheduleRun) error
	FinishScheduleRun(ctx context.Context, id domain.ID, status domain.ScheduleRunStatus, errText string, finishedAt time.Time) error
	ListScheduleRuns(ctx context.Context, scheduleID domain.ID, limit int) ([]domain.ScheduleRun, error)
}

func (r *Repository) CreateSchedule(ctx context.Context, schedule domain.Schedule) error {
	if strings.TrimSpace(string(schedule.ID)) == "" || strings.TrimSpace(schedule.Name) == "" || strings.TrimSpace(string(schedule.PekerjaanID)) == "" || strings.TrimSpace(string(schedule.AgenID)) == "" || strings.TrimSpace(schedule.Program) == "" || schedule.Interval <= 0 || schedule.NextRunAt.IsZero() {
		return fmt.Errorf("jadwal: %w", ErrInvalid)
	}
	if schedule.Status != domain.ScheduleEnabled && schedule.Status != domain.ScheduleDisabled {
		return fmt.Errorf("jadwal: status %q: %w", schedule.Status, ErrInvalid)
	}
	if schedule.RetryLimit < 0 || schedule.RetryLimit > 10 {
		return fmt.Errorf("jadwal: retry limit: %w", ErrInvalid)
	}
	args, err := json.Marshal(schedule.Arguments)
	if err != nil {
		return fmt.Errorf("jadwal: argumen: %w", err)
	}
	created, updated := timestamps(schedule.CreatedAt, schedule.UpdatedAt)
	_, err = r.db.ExecContext(ctx, `INSERT INTO schedule (id, name, pekerjaan_id, tugas_id, agen_id, program, arguments, interval_seconds, next_run_at, status, retry_limit, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, schedule.ID, schedule.Name, schedule.PekerjaanID, schedule.TugasID, schedule.AgenID, schedule.Program, string(args), int64(schedule.Interval/time.Second), schedule.NextRunAt.UTC().Format(time.RFC3339Nano), schedule.Status, schedule.RetryLimit, created, updated)
	if err != nil {
		return fmt.Errorf("buat jadwal: %w", err)
	}
	return nil
}

func (r *Repository) GetSchedule(ctx context.Context, id domain.ID) (domain.Schedule, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Schedule{}, fmt.Errorf("jadwal: %w", ErrInvalid)
	}
	return r.scanSchedule(r.db.QueryRowContext(ctx, `SELECT id, name, pekerjaan_id, tugas_id, agen_id, program, arguments, interval_seconds, next_run_at, status, retry_limit, created_at, updated_at FROM schedule WHERE id = ?`, id))
}

func (r *Repository) ListSchedules(ctx context.Context) ([]domain.Schedule, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, pekerjaan_id, tugas_id, agen_id, program, arguments, interval_seconds, next_run_at, status, retry_limit, created_at, updated_at FROM schedule ORDER BY next_run_at, id`)
	if err != nil {
		return nil, fmt.Errorf("daftar jadwal: %w", err)
	}
	defer rows.Close()
	var result []domain.Schedule
	for rows.Next() {
		schedule, err := scanScheduleRow(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, schedule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca jadwal: %w", err)
	}
	return result, nil
}

func (r *Repository) UpdateScheduleRun(ctx context.Context, id domain.ID, nextRunAt time.Time) error {
	result, err := r.db.ExecContext(ctx, `UPDATE schedule SET next_run_at = ?, updated_at = ? WHERE id = ?`, nextRunAt.UTC().Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return fmt.Errorf("perbarui jadwal: %w", err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) CreateScheduleRun(ctx context.Context, run domain.ScheduleRun) error {
	if strings.TrimSpace(string(run.ID)) == "" || strings.TrimSpace(string(run.ScheduleID)) == "" || run.Attempt < 1 {
		return fmt.Errorf("riwayat jadwal: %w", ErrInvalid)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO schedule_run (id, schedule_id, status, attempt, started_at, finished_at, error) VALUES (?, ?, ?, ?, ?, ?, ?)`, run.ID, run.ScheduleID, run.Status, run.Attempt, run.StartedAt.UTC().Format(time.RFC3339Nano), nullableTime(run.FinishedAt), run.Error)
	if err != nil {
		return fmt.Errorf("buat riwayat jadwal: %w", err)
	}
	return nil
}

func (r *Repository) FinishScheduleRun(ctx context.Context, id domain.ID, status domain.ScheduleRunStatus, errText string, finishedAt time.Time) error {
	result, err := r.db.ExecContext(ctx, `UPDATE schedule_run SET status = ?, error = ?, finished_at = ? WHERE id = ?`, status, errText, finishedAt.UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return fmt.Errorf("selesaikan riwayat jadwal: %w", err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ListScheduleRuns(ctx context.Context, scheduleID domain.ID, limit int) ([]domain.ScheduleRun, error) {
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("riwayat jadwal: limit: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, schedule_id, status, attempt, started_at, finished_at, error FROM schedule_run WHERE schedule_id = ? ORDER BY started_at DESC, id DESC LIMIT ?`, scheduleID, limit)
	if err != nil {
		return nil, fmt.Errorf("daftar riwayat jadwal: %w", err)
	}
	defer rows.Close()
	var result []domain.ScheduleRun
	for rows.Next() {
		var run domain.ScheduleRun
		var started, finished sql.NullString
		if err := rows.Scan(&run.ID, &run.ScheduleID, &run.Status, &run.Attempt, &started, &finished, &run.Error); err != nil {
			return nil, fmt.Errorf("baca riwayat jadwal: %w", err)
		}
		var err error
		run.StartedAt, err = parseTime(started.String)
		if err != nil { return nil, err }
		if finished.Valid && finished.String != "" { value, err := parseTime(finished.String); if err != nil { return nil, err }; run.FinishedAt = &value }
		result = append(result, run)
	}
	return result, rows.Err()
}

type rowScanner interface { Scan(...any) error }

func scanScheduleRow(row rowScanner) (domain.Schedule, error) {
	var s domain.Schedule
	var tugasID sql.NullString
	var args, nextRun, created, updated string
	var intervalSeconds int64
	if err := row.Scan(&s.ID, &s.Name, &s.PekerjaanID, &tugasID, &s.AgenID, &s.Program, &args, &intervalSeconds, &nextRun, &s.Status, &s.RetryLimit, &created, &updated); err != nil {
		if err == sql.ErrNoRows { return domain.Schedule{}, ErrNotFound }
		return domain.Schedule{}, fmt.Errorf("baca jadwal: %w", err)
	}
	if tugasID.Valid { id := domain.ID(tugasID.String); s.TugasID = &id }
	if err := json.Unmarshal([]byte(args), &s.Arguments); err != nil { return domain.Schedule{}, fmt.Errorf("jadwal argumen rusak: %w", err) }
	var err error
	s.Interval = time.Duration(intervalSeconds) * time.Second
	s.NextRunAt, err = parseTime(nextRun); if err != nil { return domain.Schedule{}, err }
	s.CreatedAt, err = parseTime(created); if err != nil { return domain.Schedule{}, err }
	s.UpdatedAt, err = parseTime(updated); if err != nil { return domain.Schedule{}, err }
	return s, nil
}

func (r *Repository) scanSchedule(row *sql.Row) (domain.Schedule, error) { return scanScheduleRow(row) }

func nullableTime(value *time.Time) any { if value == nil { return nil }; return value.UTC().Format(time.RFC3339Nano) }
