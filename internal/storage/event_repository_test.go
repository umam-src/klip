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
	event := domain.Event{
		ID:          "event-1",
		PekerjaanID: "pekerjaan-1",
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
	if !events[0].CreatedAt.Equal(created) {
		t.Fatalf("CreatedAt = %v, want %v", events[0].CreatedAt, created)
	}
}

func TestExecutionLifecycleRecordsEvents(t *testing.T) {
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
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Pekerjaan", Status: domain.StatusReady}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Tugas", Status: domain.StatusReady}); err != nil {
		t.Fatal(err)
	}

	tugasID := domain.ID("tugas-1")
	sesi := domain.Sesi{ID: "sesi-1", PekerjaanID: "pekerjaan-1", AgenID: "agen-1", Status: domain.StatusRunning}
	run := domain.Run{ID: "run-1", PekerjaanID: "pekerjaan-1", TugasID: &tugasID, AgenID: "agen-1", Status: domain.StatusRunning, Program: "echo"}
	if _, err := repo.StartExecution(ctx, sesi, run); err != nil {
		t.Fatal(err)
	}

	events, err := repo.ListEventsByPekerjaan(ctx, "pekerjaan-1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].Type != domain.EventExecutionStarted {
		t.Fatalf("start events = %+v, want one execution.started event", events)
	}

	finished := time.Now().UTC()
	if err := repo.FinalizeExecution(ctx, run.ID, sesi.ID, &tugasID, domain.StatusCompleted, intPtr(0), "ok", "", finished); err != nil {
		t.Fatal(err)
	}

	events, err = repo.ListEventsByPekerjaan(ctx, "pekerjaan-1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 2 {
		t.Fatalf("events = %+v, want 2 events", events)
	}
	if events[0].Type != domain.EventExecutionCompleted || events[1].Type != domain.EventExecutionStarted {
		t.Fatalf("event types = %q, %q, want completed then started", events[0].Type, events[1].Type)
	}
	if events[0].RunID == nil || *events[0].RunID != run.ID {
		t.Fatalf("terminal event RunID = %v, want %q", events[0].RunID, run.ID)
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
