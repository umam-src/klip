package storage

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func TestCreateScheduleRejectsCrossWorkspaceAgent(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	seedScheduleBoundaryFixtures(t, db)

	repo := &Repository{db: db}
	err = repo.CreateSchedule(ctx, validSchedule("cross-agent", "ruang-1", "proyek-1", nil, "agen-2"))
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateSchedule() error = %v, want ErrInvalid", err)
	}
}

func TestCreateScheduleRejectsTaskFromAnotherProyek(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	seedScheduleBoundaryFixtures(t, db)

	tugasID := domain.ID("tugas-2")
	repo := &Repository{db: db}
	err = repo.CreateSchedule(ctx, validSchedule("cross-task", "ruang-1", "proyek-1", &tugasID, "agen-1"))
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateSchedule() error = %v, want ErrInvalid", err)
	}
}

func TestCreateScheduleAcceptsSameWorkspaceRelations(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	seedScheduleBoundaryFixtures(t, db)

	tugasID := domain.ID("tugas-1")
	repo := &Repository{db: db}
	if err := repo.CreateSchedule(ctx, validSchedule("valid", "ruang-1", "proyek-1", &tugasID, "agen-1")); err != nil {
		t.Fatalf("CreateSchedule() error = %v", err)
	}
	if _, err := repo.GetSchedule(ctx, "valid"); err != nil {
		t.Fatalf("GetSchedule() error = %v", err)
	}
}

func validSchedule(id, ruangID, proyekID string, tugasID *domain.ID, agenID string) domain.Schedule {
	return domain.Schedule{
		ID:         domain.ID(id),
		Name:       "uji",
		RuangID:    domain.ID(ruangID),
		ProyekID:   domain.ID(proyekID),
		TugasID:    tugasID,
		AgenID:     domain.ID(agenID),
		Program:    "true",
		Interval:   time.Minute,
		NextRunAt:  time.Now().UTC().Add(time.Minute),
		Status:     domain.ScheduleEnabled,
	}
}

func seedScheduleBoundaryFixtures(t *testing.T, db *sql.DB) {
	t.Helper()
	now := time.Now().UTC().Format(time.RFC3339Nano)
	statements := []struct {
		query string
		args  []any
	}{
		{`INSERT INTO ruang (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`, []any{"ruang-1", "Ruang 1", now, now}},
		{`INSERT INTO ruang (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`, []any{"ruang-2", "Ruang 2", now, now}},
		{`INSERT INTO agen (id, ruang_id, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, []any{"agen-1", "ruang-1", "Agen 1", now, now}},
		{`INSERT INTO agen (id, ruang_id, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, []any{"agen-2", "ruang-2", "Agen 2", now, now}},
		{`INSERT INTO goal (id, ruang_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, []any{"goal-1", "ruang-1", "Goal 1", "", "active", now, now}},
		{`INSERT INTO goal (id, ruang_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, []any{"goal-2", "ruang-1", "Goal 2", "", "active", now, now}},
		{`INSERT INTO proyek (id, ruang_id, goal_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, []any{"proyek-1", "ruang-1", "goal-1", "Proyek 1", "", "ready", now, now}},
		{`INSERT INTO proyek (id, ruang_id, goal_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, []any{"proyek-2", "ruang-1", "goal-2", "Proyek 2", "", "ready", now, now}},
		{`INSERT INTO tugas (id, ruang_id, proyek_id, title, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, []any{"tugas-1", "ruang-1", "proyek-1", "Tugas 1", "ready", now, now}},
		{`INSERT INTO tugas (id, ruang_id, proyek_id, title, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, []any{"tugas-2", "ruang-1", "proyek-2", "Tugas 2", "ready", now, now}},
	}
	for _, statement := range statements {
		if _, err := db.Exec(statement.query, statement.args...); err != nil {
			t.Fatalf("seed fixture: %v", err)
		}
	}
}
