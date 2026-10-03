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
	ListEventsByPekerjaan(ctx context.Context, pekerjaanID domain.ID, limit int) ([]domain.Event, error)
}

func (r *Repository) AppendEvent(ctx context.Context, event domain.Event) error {
	if err := r.validateEventContext(ctx, event); err != nil {
		return err
	}
	return appendEvent(ctx, r.db, event)
}

func (r *Repository) validateEventContext(ctx context.Context, event domain.Event) error {
	if strings.TrimSpace(string(event.PekerjaanID)) == "" {
		return fmt.Errorf("event: pekerjaan wajib diisi: %w", ErrInvalid)
	}

	pekerjaan, err := r.GetPekerjaan(ctx, event.PekerjaanID)
	if err != nil {
		return err
	}

	if event.TugasID != nil {
		tugas, err := r.GetTugas(ctx, *event.TugasID)
		if err != nil {
			return err
		}
		if tugas.PekerjaanID != event.PekerjaanID {
			return fmt.Errorf("event: tugas tidak sesuai dengan pekerjaan: %w", ErrInvalid)
		}
	}

	if event.AgenID != nil {
		agent, err := r.GetAgen(ctx, *event.AgenID)
		if err != nil {
			return err
		}
		if agent.RuangID != pekerjaan.RuangID {
			return fmt.Errorf("event: agen berada di ruang kerja berbeda: %w", ErrInvalid)
		}
	}

	if event.SesiID != nil {
		sesi, err := r.GetSesi(ctx, *event.SesiID)
		if err != nil {
			return err
		}
		if sesi.PekerjaanID != event.PekerjaanID {
			return fmt.Errorf("event: sesi tidak sesuai dengan pekerjaan: %w", ErrInvalid)
		}
		if event.AgenID != nil && sesi.AgenID != *event.AgenID {
			return fmt.Errorf("event: sesi tidak sesuai dengan agen: %w", ErrInvalid)
		}
	}

	if event.RunID != nil {
		run, err := r.GetRun(ctx, *event.RunID)
		if err != nil {
			return err
		}
		if run.PekerjaanID != event.PekerjaanID {
			return fmt.Errorf("event: run tidak sesuai dengan pekerjaan: %w", ErrInvalid)
		}
		if event.TugasID != nil && (run.TugasID == nil || *run.TugasID != *event.TugasID) {
			return fmt.Errorf("event: run tidak sesuai dengan tugas: %w", ErrInvalid)
		}
		if event.AgenID != nil && run.AgenID != *event.AgenID {
			return fmt.Errorf("event: run tidak sesuai dengan agen: %w", ErrInvalid)
		}
	}

	return nil
}

func (r *Repository) ListEventsByPekerjaan(ctx context.Context, pekerjaanID domain.ID, limit int) ([]domain.Event, error) {
	if strings.TrimSpace(string(pekerjaanID)) == "" || limit < 1 || limit > 1000 {
		return nil, fmt.Errorf("event: parameter tidak valid: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, pekerjaan_id, tugas_id, sesi_id, run_id, agen_id, type, message, created_at
		FROM event
		WHERE pekerjaan_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?`, pekerjaanID, limit)
	if err != nil {
		return nil, fmt.Errorf("daftar event: %w", err)
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
		return nil, fmt.Errorf("baca daftar event: %w", err)
	}
	return result, nil
}

func appendEvent(ctx context.Context, exec interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}, event domain.Event) error {
	if strings.TrimSpace(string(event.PekerjaanID)) == "" || strings.TrimSpace(event.Type) == "" {
		return fmt.Errorf("event: pekerjaan dan type wajib diisi: %w", ErrInvalid)
	}
	if strings.ContainsRune(event.Type, '\x00') || strings.ContainsRune(event.Message, '\x00') {
		return fmt.Errorf("event: input mengandung karakter NUL: %w", ErrInvalid)
	}
	if event.ID == "" {
		event.ID = newEventID()
	}
	created := event.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}

	var tugasID, sesiID, runID, agenID any
	if event.TugasID != nil {
		tugasID = string(*event.TugasID)
	}
	if event.SesiID != nil {
		sesiID = string(*event.SesiID)
	}
	if event.RunID != nil {
		runID = string(*event.RunID)
	}
	if event.AgenID != nil {
		agentID := string(*event.AgenID)
		agenID = agentID
	}
	if _, err := exec.ExecContext(ctx, `
		INSERT INTO event (id, pekerjaan_id, tugas_id, sesi_id, run_id, agen_id, type, message, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.ID, event.PekerjaanID, tugasID, sesiID, runID, agenID, event.Type, event.Message,
		created.UTC().Format(time.RFC3339Nano)); err != nil {
		return fmt.Errorf("simpan event: %w", err)
	}
	return nil
}

type eventScanner interface {
	Scan(...any) error
}

func scanEvent(scanner eventScanner) (domain.Event, error) {
	var event domain.Event
	var tugasID, sesiID, runID, agenID sql.NullString
	var created string
	if err := scanner.Scan(&event.ID, &event.PekerjaanID, &tugasID, &sesiID, &runID, &agenID, &event.Type, &event.Message, &created); err != nil {
		return domain.Event{}, fmt.Errorf("baca event: %w", err)
	}
	if tugasID.Valid {
		value := domain.ID(tugasID.String)
		event.TugasID = &value
	}
	if sesiID.Valid {
		value := domain.ID(sesiID.String)
		event.SesiID = &value
	}
	if runID.Valid {
		value := domain.ID(runID.String)
		event.RunID = &value
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
