package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func TestAppendAndListEvents(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Pekerjaan", Status: domain.StatusReady}); err != nil {
		t.Fatal(err)
	}

	created := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	runID := domain.ID("run-1")
	event := domain.Event{
		ID:          "event-1",
		PekerjaanID: "pekerjaan-1",
		RunID:       &runID,
		Type:        domain.EventExecutionStarted,
		Message:     "Eksekusi dimulai",
		CreatedAt:   created,
	}
	if err := repo.AppendEvent(ctx, event); err != nil {
		t.Fatal(err)
	}

	events, err := repo.ListEventsByPekerjaan(ctx, "pekerjaan-1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	if events[0].ID != event.ID || events[0].Type != event.Type || events[0].Message != event.Message {
		t.Fatalf("event = %+v, want %+v", events[0], event)
	}
	if events[0].RunID == nil || *events[0].RunID != runID {
		t.Fatalf("RunID = %v, want %q", events[0].RunID, runID)
	}
	if !events[0].CreatedAt.Equal(created) {
		t.Fatalf("CreatedAt = %v, want %v", events[0].CreatedAt, created)
	}
}

func TestAppendEventRejectsInvalidInput(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRepository(db)
	err = repo.AppendEvent(ctx, domain.Event{Type: domain.EventExecutionStarted})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("AppendEvent() error = %v, want ErrInvalid", err)
	}

	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Pekerjaan", Status: domain.StatusReady}); err != nil {
		t.Fatal(err)
	}
	if err := repo.AppendEvent(ctx, domain.Event{PekerjaanID: "pekerjaan-1", Type: "bad\x00type"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("AppendEvent() NUL error = %v, want ErrInvalid", err)
	}
	if _, err := repo.ListEventsByPekerjaan(ctx, "pekerjaan-1", 0); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ListEventsByPekerjaan() error = %v, want ErrInvalid", err)
	}
}
