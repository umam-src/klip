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
	if strings.TrimSpace(string(schedule.ID)) == "" ||
		strings.TrimSpace(schedule.Name) == "" ||
		strings.TrimSpace(string(schedule.RuangID)) == "" ||
		strings.TrimSpace(string(schedule.ProyekID)) == "" ||
		strings.TrimSpace(string(schedule.AgenID)) == "" ||
		strings.TrimSpace(schedule.Program) == "" ||
		schedule.Interval <= 0 || schedule.NextRunAt.IsZero() {
		return fmt.Errorf("jadwal: %w", ErrInvalid)
	}
	if schedule.Status != domain.ScheduleEnabled && schedule.Status != domain.ScheduleDisabled {
		return fmt.Errorf("jadwal: status %q: %w", schedule.Status, ErrInvalid)
	}
	if schedule.RetryLimit < 0 || schedule.RetryLimit > 10 {
		return fmt.Errorf("jadwal: retry limit: %w", ErrInvalid)
	}

	var proyekRuangID, agenRuangID string
	if err := r.db.QueryRowContext(ctx, `SELECT ruang_id FROM proyek WHERE id = ?`, schedule.ProyekID).Scan(&proyekRuangID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("jadwal: proyek tidak ditemukan: %w", ErrInvalid)
		}
		return fmt.Errorf("cek ruang proyek jadwal: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT ruang_id FROM agen WHERE id = ?`, schedule.AgenID).Scan(&agenRuangID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("jadwal: agen tidak ditemukan: %w", ErrInvalid)
		}
		return fmt.Errorf("cek ruang agen jadwal: %w", err)
	}
	if proyekRuangID != string(schedule.RuangID) || agenRuangID != string(schedule.RuangID) {
		return fmt.Errorf("jadwal: konteks ruang tidak konsisten: %w", ErrInvalid)
	}
	if schedule.TugasID != nil {
		value := strings.TrimSpace(string(*schedule.TugasID))
		if value == "" {
			return fmt.Errorf("jadwal: tugas tidak valid: %w", ErrInvalid)
		}
		var tugasProyekID, tugasRuangID string
		if err := r.db.QueryRowContext(ctx, `SELECT proyek_id, ruang_id FROM tugas WHERE id = ?`, value).Scan(&tugasProyekID, &tugasRuangID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("jadwal: tugas tidak ditemukan: %w", ErrInvalid)
			}
			return fmt.Errorf("cek tugas jadwal: %w", err)
		}
		if tugasProyekID != string(schedule.ProyekID) || tugasRuangID != string(schedule.RuangID) {
			return fmt.Errorf("jadwal: tugas tidak sesuai dengan proyek atau ruang: %w", ErrInvalid)
		}
	}

	args, err := json.Marshal(schedule.Arguments)
	if err != nil {
		return fmt.Errorf("jadwal: argumen: %w", err)
	}
	created, updated := timestamps(schedule.CreatedAt, schedule.UpdatedAt)
	_, err = r.db.ExecContext(ctx, `INSERT INTO jadwal (id, name, ruang_id, proyek_id, tugas_id, agen_id, program, arguments, interval_seconds, next_run_at, status, retry_limit, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		schedule.ID, schedule.Name, schedule.RuangID, schedule.ProyekID, schedule.TugasID, schedule.AgenID,
		schedule.Program, string(args), int64(schedule.Interval/time.Second),
		schedule.NextRunAt.UTC().Format(time.RFC3339Nano), schedule.Status, schedule.RetryLimit, created, updated)
	if err != nil {
		return fmt.Errorf("buat jadwal: %w", err)
	}
	return nil
}

func (r *Repository) GetSchedule(ctx context.Context, id domain.ID) (domain.Schedule, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Schedule{}, fmt.Errorf("jadwal: %w", ErrInvalid)
	}
	return r.scanSchedule(r.db.QueryRowContext(ctx, `SELECT id, name, ruang_id, proyek_id, tugas_id, agen_id, program, arguments, interval_seconds, next_run_at, status, retry_limit, created_at, updated_at FROM jadwal WHERE id = ?`, id))
}

func (r *Repository) ListSchedules(ctx context.Context) ([]domain.Schedule, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, ruang_id, proyek_id, tugas_id, agen_id, program, arguments, interval_seconds, next_run_at, status, retry_limit, created_at, updated_at FROM jadwal ORDER BY next_run_at, id`)
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
	result, err := r.db.ExecContext(ctx, `UPDATE jadwal SET next_run_at = ?, updated_at = ? WHERE id = ?`, nextRunAt.UTC().Format(time.RFC3339Nano), time.Now().UTC().Format(time.RFC3339Nano), id)
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
		return fmt.Errorf("jadwal eksekusi: %w", ErrInvalid)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO jadwal_eksekusi (id, jadwal_id, status, attempt, started_at, finished_at, error) VALUES (?, ?, ?, ?, ?, ?, ?)`, run.ID, run.ScheduleID, run.Status, run.Attempt, run.StartedAt.UTC().Format(time.RFC3339Nano), nullableTime(run.FinishedAt), run.Error)
	if err != nil {
		return fmt.Errorf("buat jadwal eksekusi: %w", err)
	}
	return nil
}

func (r *Repository) FinishScheduleRun(ctx context.Context, id domain.ID, status domain.ScheduleRunStatus, errText string, finishedAt time.Time) error {
	result, err := r.db.ExecContext(ctx, `UPDATE jadwal_eksekusi SET status = ?, error = ?, finished_at = ? WHERE id = ?`, status, errText, finishedAt.UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return fmt.Errorf("selesaikan jadwal eksekusi: %w", err)
	}
	if n, _ := result.RowsAffected(); n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ListScheduleRuns(ctx context.Context, scheduleID domain.ID, limit int) ([]domain.ScheduleRun, error) {
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("jadwal eksekusi: limit: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, jadwal_id, status, attempt, started_at, finished_at, error FROM jadwal_eksekusi WHERE jadwal_id = ? ORDER BY started_at DESC, id DESC LIMIT ?`, scheduleID, limit)
	if err != nil {
		return nil, fmt.Errorf("daftar jadwal eksekusi: %w", err)
	}
	defer rows.Close()
	var result []domain.ScheduleRun
	for rows.Next() {
		var run domain.ScheduleRun
		var started, finished sql.NullString
		if err := rows.Scan(&run.ID, &run.ScheduleID, &run.Status, &run.Attempt, &started, &finished, &run.Error); err != nil {
			return nil, fmt.Errorf("baca jadwal eksekusi: %w", err)
		}
		var err error
		run.StartedAt, err = parseTime(started.String)
		if err != nil {
			return nil, err
		}
		if finished.Valid && finished.String != "" {
			value, err := parseTime(finished.String)
			if err != nil {
				return nil, err
			}
			run.FinishedAt = &value
		}
		result = append(result, run)
	}
	return result, rows.Err()
}

type rowScanner interface {
	Scan(...any) error
}

func scanScheduleRow(row rowScanner) (domain.Schedule, error) {
	var schedule domain.Schedule
	var tugasID sql.NullString
	var args, nextRun, created, updated string
	var intervalSeconds int64
	if err := row.Scan(&schedule.ID, &schedule.Name, &schedule.RuangID, &schedule.ProyekID, &tugasID, &schedule.AgenID, &schedule.Program, &args, &intervalSeconds, &nextRun, &schedule.Status, &schedule.RetryLimit, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Schedule{}, ErrNotFound
		}
		return domain.Schedule{}, fmt.Errorf("baca jadwal: %w", err)
	}
	if tugasID.Valid {
		id := domain.ID(tugasID.String)
		schedule.TugasID = &id
	}
	if err := json.Unmarshal([]byte(args), &schedule.Arguments); err != nil {
		return domain.Schedule{}, fmt.Errorf("jadwal argumen rusak: %w", err)
	}
	var err error
	schedule.Interval = time.Duration(intervalSeconds) * time.Second
	schedule.NextRunAt, err = parseTime(nextRun)
	if err != nil {
		return domain.Schedule{}, err
	}
	schedule.CreatedAt, err = parseTime(created)
	if err != nil {
		return domain.Schedule{}, err
	}
	schedule.UpdatedAt, err = parseTime(updated)
	if err != nil {
		return domain.Schedule{}, err
	}
	return schedule, nil
}

func (r *Repository) scanSchedule(row *sql.Row) (domain.Schedule, error) {
	return scanScheduleRow(row)
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}
