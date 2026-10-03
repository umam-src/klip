package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func buatKonteksPeristiwa(t *testing.T) (context.Context, *Repository) {
	t.Helper()
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	repo := NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateGoal(ctx, domain.Goal{ID: "goal-1", RuangID: "ruang-1", Title: "Goal", Status: domain.GoalStatusActive}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateProyek(ctx, domain.Proyek{ID: "proyek-1", RuangID: "ruang-1", GoalID: "goal-1", Title: "Proyek", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}
	return ctx, repo
}

func TestAppendAndListEvents(t *testing.T) {
	ctx, repo := buatKonteksPeristiwa(t)
	created := time.Date(2026, 10, 2, 12, 0, 0, 0, time.UTC)
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateEksekusi(ctx, domain.Eksekusi{ID: "eksekusi-1", RuangID: "ruang-1", ProyekID: "proyek-1", AgenID: "agen-1", Status: domain.StatusRunning, Program: "echo", StartedAt: created}); err != nil {
		t.Fatal(err)
	}
	event := domain.Event{ID: "event-1", RuangID: "ruang-1", ExecutionID: idPtr("eksekusi-1"), ProyekID: "proyek-1", Type: domain.EventExecutionStarted, Message: "Eksekusi dimulai", CreatedAt: created}
	if err := repo.AppendEvent(ctx, event); err != nil {
		t.Fatal(err)
	}
	events, err := repo.ListEventsByProyek(ctx, "proyek-1", 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].ID != event.ID || events[0].Type != event.Type || events[0].Message != event.Message || !events[0].CreatedAt.Equal(created) {
		t.Fatalf("events = %+v, want %+v", events, event)
	}
}

func TestAppendEventRejectsInvalidInput(t *testing.T) {
	ctx, repo := buatKonteksPeristiwa(t)
	if err := repo.AppendEvent(ctx, domain.Event{Type: domain.EventExecutionStarted}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("AppendEvent() error = %v, want ErrInvalid", err)
	}
	if err := repo.AppendEvent(ctx, domain.Event{RuangID: "ruang-1", ProyekID: "proyek-1", Type: "bad\x00type"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("AppendEvent() NUL error = %v, want ErrInvalid", err)
	}
	if _, err := repo.ListEventsByProyek(ctx, "proyek-1", 0); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ListEventsByProyek() error = %v, want ErrInvalid", err)
	}
}

func TestAppendEventRejectsCrossWorkspaceContext(t *testing.T) {
	ctx, repo := buatKonteksPeristiwa(t)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-2", Name: "Ruang 2"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen 1"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-2", RuangID: "ruang-2", Name: "Agen 2"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateGoal(ctx, domain.Goal{ID: "goal-2", RuangID: "ruang-2", Title: "Goal 2", Status: domain.GoalStatusActive}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateProyek(ctx, domain.Proyek{ID: "proyek-2", RuangID: "ruang-2", GoalID: "goal-2", Title: "Proyek 2", Status: domain.StatusDraft}); err != nil {
		t.Fatal(err)
	}
	tugas1 := domain.ID("tugas-1")
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: tugas1, RuangID: "ruang-1", ProyekID: "proyek-1", Title: "Tugas 1", Status: domain.StatusReady}); err != nil {
		t.Fatal(err)
	}
	tugas2 := domain.ID("tugas-2")
	if err := repo.CreateTugasNative(ctx, domain.Tugas{ID: tugas2, RuangID: "ruang-2", ProyekID: "proyek-2", Title: "Tugas 2", Status: domain.StatusReady}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateEksekusi(ctx, domain.Eksekusi{ID: "eksekusi-1", RuangID: "ruang-1", ProyekID: "proyek-1", TugasID: idPtr("tugas-1"), AgenID: "agen-1", Status: domain.StatusRunning, Program: "echo", StartedAt: time.Now().UTC()}); err != nil {
		t.Fatal(err)
	}
	cases := []domain.Event{
		{RuangID: "ruang-1", ProyekID: "proyek-1", AgenID: idPtr("agen-2"), Type: domain.EventExecutionStarted},
		{RuangID: "ruang-1", ProyekID: "proyek-1", TugasID: idPtr("tugas-2"), Type: domain.EventExecutionStarted},
	}
	for _, event := range cases {
		if err := repo.AppendEvent(ctx, event); !errors.Is(err, ErrInvalid) {
			t.Errorf("AppendEvent() error = %v, want ErrInvalid", err)
		}
	}
	if err := repo.AppendEvent(ctx, domain.Event{RuangID: "ruang-1", ExecutionID: idPtr("eksekusi-1"), ProyekID: "proyek-1", TugasID: idPtr("tugas-1"), AgenID: idPtr("agen-1"), Type: domain.EventExecutionStarted}); err != nil {
		t.Fatalf("AppendEvent() valid context error = %v", err)
	}
}

func idPtr(value string) *domain.ID {
	id := domain.ID(value)
	return &id
}
