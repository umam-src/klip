package storage

import (
	"context"
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
	err = repo.CreateSchedule(ctx, validSchedule("cross-agent", "pekerjaan-1", nil, "agen-2"))
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateSchedule() error = %v, want ErrInvalid", err)
	}
}

func TestCreateScheduleRejectsTaskFromAnotherPekerjaan(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	seedScheduleBoundaryFixtures(t, db)
	tugasID := domain.ID("tugas-2")
	repo := &Repository{db: db}
	err = repo.CreateSchedule(ctx, validSchedule("cross-task", "pekerjaan-1", &tugasID, "agen-1"))
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
	if err := repo.CreateSchedule(ctx, validSchedule("valid", "pekerjaan-1", &tugasID, "agen-1")); err != nil {
		t.Fatalf("CreateSchedule() error = %v", err)
	}
	if _, err := repo.GetSchedule(ctx, "valid"); err != nil {
		t.Fatalf("GetSchedule() error = %v", err)
	}
}

func validSchedule(id, pekerjaanID string, tugasID *domain.ID, agenID string) domain.Schedule {
	return domain.Schedule{
		ID:          domain.ID(id),
		Name:        "uji",
		PekerjaanID: domain.ID(pekerjaanID),
		TugasID:     tugasID,
		AgenID:      domain.ID(agenID),
		Program:     "true",
		Interval:    time.Minute,
		NextRunAt:   time.Now().UTC().Add(time.Minute),
		Status:      domain.ScheduleEnabled,
	}
}

func seedScheduleBoundaryFixtures(t *testing.T, db interface {
	ExecContext(context.Context, string, ...any) (interface{ LastInsertId() (int64, error); RowsAffected() (int64, error) }, error)
}) {
	t.Helper()
}
