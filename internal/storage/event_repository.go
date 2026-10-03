package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

type EventRepository interface {
	AppendEvent(ctx context.Context, event domain.Event) error
	ListEventsByProyek(ctx context.Context, proyekID domain.ID, limit int) ([]domain.Event, error)
}

func (r *Repository) AppendEvent(ctx context.Context, event domain.Event) error {
	if err := r.validateEventContext(ctx, event); err != nil {
		return err
	}
	return appendEvent(ctx, r.db, event)
}

func (r *Repository) validateEventContext(ctx context.Context, event domain.Event) error {
	if strings.TrimSpace(string(event.RuangID)) == "" || strings.TrimSpace(string(event.ProyekID)) == "" {
		return fmt.Errorf("peristiwa: ruang dan proyek wajib diisi: %w", ErrInvalid)
	}
	proyek, err := r.GetProyek(ctx, event.ProyekID)
	if err != nil {
		return err
	}
	if proyek.RuangID != event.RuangID {
		return fmt.Errorf("peristiwa: proyek tidak sesuai dengan ruang: %w", ErrInvalid)
	}
	if event.TugasID != nil {
		tugas, err := r.GetTugasNative(ctx, *event.TugasID)
		if err != nil {
			return err
		}
		if tugas.ProyekID != event.ProyekID || tugas.RuangID != event.RuangID {
			return fmt.Errorf("peristiwa: tugas tidak sesuai dengan proyek atau ruang: %w", ErrInvalid)
		}
	}
	if event.AgenID != nil {
		agent, err := r.GetAgen(ctx, *event.AgenID)
		if err != nil {
			return err
		}
		if agent.RuangID != event.RuangID {
			return fmt.Errorf("peristiwa: agen berada di ruang kerja berbeda: %w", ErrInvalid)
		}
	}
	if event.ExecutionID != nil {
		execution, err := r.GetEksekusi(ctx, *event.ExecutionID)
		if err != nil {
			return err
		}
		if execution.RuangID != event.RuangID || execution.ProyekID != event.ProyekID {
			return fmt.Errorf("peristiwa: eksekusi tidak sesuai dengan konteks: %w", ErrInvalid)
		}
		if event.TugasID != nil && (execution.TugasID == nil || *execution.TugasID != *event.TugasID) {
			return fmt.Errorf("peristiwa: eksekusi tidak sesuai dengan tugas: %w", ErrInvalid)
		}
		if event.AgenID != nil && execution.AgenID != *event.AgenID {
			return fmt.Errorf("peristiwa: eksekusi tidak sesuai dengan agen: %w", ErrInvalid)
		}
	}
	return nil
}

func (r *Repository) ListEventsByProyek(ctx context.Context, proyekID domain.ID, limit int) ([]domain.Event, error) {
	if strings.TrimSpace(string(proyekID)) == "" || limit < 1 || limit > 1000 {
		return nil, fmt.Errorf("peristiwa: parameter tidak valid: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, ruang_id, execution_id, proyek_id, tugas_id, agen_id, type, message, created_at FROM peristiwa WHERE proyek_id = ? ORDER BY created_at DESC, id DESC LIMIT ?`, proyekID, limit)
	if err != nil {
		return nil, fmt.Errorf("daftar peristiwa: %w", err)
	}
	defer rows.Close()
	var result []domain.Event
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar peristiwa: %w", err)
	}
	return result, nil
}

func appendEvent(ctx context.Context, exec interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, event domain.Event) error {
	if strings.TrimSpace(string(event.RuangID)) == "" || strings.TrimSpace(string(event.ProyekID)) == "" || strings.TrimSpace(event.Type) == "" {
		return fmt.Errorf("peristiwa: ruang, proyek, dan type wajib diisi: %w", ErrInvalid)
	}
	if strings.ContainsRune(event.Type, '\x00') || strings.ContainsRune(event.Message, '\x00') {
		return fmt.Errorf("peristiwa: input mengandung karakter NUL: %w", ErrInvalid)
	}
	if event.ID == "" {
		event.ID = newEventID()
	}
	created := event.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}
	var executionID, tugasID, agenID any
	if event.ExecutionID != nil {
		executionID = string(*event.ExecutionID)
	}
	if event.TugasID != nil {
		tugasID = string(*event.TugasID)
	}
	if event.AgenID != nil {
		agentID = string(*event.AgenID)
	}
	_, err := exec.ExecContext(ctx, `INSERT INTO peristiwa (id, ruang_id, execution_id, proyek_id, tugas_id, agen_id, type, message, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, event.ID, event.RuangID, executionID, event.ProyekID, tugasID, agenID, event.Type, event.Message, created.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("simpan peristiwa: %w", err)
	}
	return nil
}

type eventScanner interface{ Scan(...any) error }

func scanEvent(scanner eventScanner) (domain.Event, error) {
	var event domain.Event
	var executionID, tugasID, agenID sql.NullString
	var created string
	if err := scanner.Scan(&event.ID, &event.RuangID, &executionID, &event.ProyekID, &tugasID, &agenID, &event.Type, &event.Message, &created); err != nil {
		return domain.Event{}, fmt.Errorf("baca peristiwa: %w", err)
	}
	if executionID.Valid {
		value := domain.ID(executionID.String)
		event.ExecutionID = &value
	}
	if tugasID.Valid {
		value := domain.ID(tugasID.String)
		event.TugasID = &value
	}
	if agenID.Valid {
		value := domain.ID(agenID.String)
		event.AgenID = &value
	}
	var err error
	if event.CreatedAt, err = parseTime(created); err != nil {
		return domain.Event{}, err
	}
	return event, nil
}

func newEventID() domain.ID {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		return domain.ID(fmt.Sprintf("event-%d", time.Now().UnixNano()))
	}
	return domain.ID("event-" + hex.EncodeToString(data[:]))
}

var _ EventRepository = (*Repository)(nil)
