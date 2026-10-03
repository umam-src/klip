package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func TestSesiRepositoryLifecycle(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	now := time.Date(2026, 10, 2, 8, 0, 0, 0, time.UTC)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-sesi", Name: "Ruang", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-sesi", RuangID: "ruang-sesi", Name: "Agen", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-sesi", RuangID: "ruang-sesi", Title: "Pekerjaan", Status: domain.StatusReady, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}

	sesi := domain.Sesi{ID: "sesi-1", PekerjaanID: "pekerjaan-sesi", AgenID: "agen-sesi", Status: domain.StatusRunning, StartedAt: now}
	if err := repo.CreateSesi(ctx, sesi); err != nil {
		t.Fatalf("CreateSesi() error = %v", err)
	}
	if err := repo.FinishSesi(ctx, sesi.ID, domain.StatusCompleted, now.Add(time.Minute)); err != nil {
		t.Fatalf("FinishSesi() error = %v", err)
	}
	if err := repo.FinishSesi(ctx, sesi.ID, domain.StatusFailed, now.Add(2*time.Minute)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("second finish error = %v, want ErrInvalid", err)
	}

	got, err := repo.GetSesi(ctx, sesi.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.StatusCompleted || got.FinishedAt == nil {
		t.Fatalf("sesi = %+v", got)
	}
}

func TestSesiRepositoryRejectsCrossWorkspaceAgent(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	now := time.Now().UTC()
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-sesi-a", Name: "A", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-sesi-b", Name: "B", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-sesi-b", RuangID: "ruang-sesi-b", Name: "Agen B", CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-sesi-a", RuangID: "ruang-sesi-a", Title: "Pekerjaan A", Status: domain.StatusReady, CreatedAt: now, UpdatedAt: now}); err != nil {
		t.Fatal(err)
	}

	err = repo.CreateSesi(ctx, domain.Sesi{ID: "sesi-cross", PekerjaanID: "pekerjaan-sesi-a", AgenID: "agen-sesi-b", Status: domain.StatusRunning, StartedAt: now})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateSesi() error = %v, want ErrInvalid", err)
	}
}

func TestSesiRepositoryRejectsInvalidLifecycle(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.CreateSesi(ctx, domain.Sesi{ID: "sesi-invalid", PekerjaanID: "pekerjaan", AgenID: "agen", Status: domain.StatusCompleted}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateSesi() error = %v, want ErrInvalid", err)
	}
	if err := repo.FinishSesi(ctx, "missing", domain.StatusCompleted, time.Now()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("FinishSesi() error = %v, want ErrNotFound", err)
	}
}
