package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func TestRepositoryCoreEntities(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	created := time.Date(2026, 10, 2, 5, 0, 0, 0, time.UTC)
	updated := created.Add(time.Minute)

	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Proyek", CreatedAt: created, UpdatedAt: updated}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen Lokal", ProviderID: "ollama", ModelID: "qwen", CreatedAt: created, UpdatedAt: updated}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Bangun inti", Status: domain.StatusReady, CreatedAt: created, UpdatedAt: updated}); err != nil {
		t.Fatalf("CreatePekerjaan() error = %v", err)
	}

	ruang, err := repo.GetRuang(ctx, "ruang-1")
	if err != nil || ruang.Name != "Proyek" || !ruang.CreatedAt.Equal(created) || !ruang.UpdatedAt.Equal(updated) {
		t.Fatalf("GetRuang() = %+v, error = %v", ruang, err)
	}

	agen, err := repo.GetAgen(ctx, "agen-1")
	if err != nil || agen.RuangID != "ruang-1" || agen.ProviderID != "ollama" {
		t.Fatalf("GetAgen() = %+v, error = %v", agen, err)
	}

	pekerjaan, err := repo.GetPekerjaan(ctx, "pekerjaan-1")
	if err != nil || pekerjaan.Status != domain.StatusReady || pekerjaan.RuangID != "ruang-1" {
		t.Fatalf("GetPekerjaan() = %+v, error = %v", pekerjaan, err)
	}

	agents, err := repo.ListAgenByRuang(ctx, "ruang-1")
	if err != nil || len(agents) != 1 || agents[0].ID != "agen-1" {
		t.Fatalf("ListAgenByRuang() = %+v, error = %v", agents, err)
	}

	jobs, err := repo.ListPekerjaanByRuang(ctx, "ruang-1")
	if err != nil || len(jobs) != 1 || jobs[0].ID != "pekerjaan-1" {
		t.Fatalf("ListPekerjaanByRuang() = %+v, error = %v", jobs, err)
	}
}

func TestRepositoryNotFoundAndValidation(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if _, err := repo.GetRuang(ctx, "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetRuang() error = %v, want ErrNotFound", err)
	}
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "", Name: "tanpa id"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateRuang() error = %v, want ErrInvalid", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", Name: "tanpa ruang"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateAgen() error = %v, want ErrInvalid", err)
	}
}

func TestRepositoryForeignKeys(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	err = repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "missing", Name: "Agen"})
	if err == nil {
		t.Fatal("CreateAgen() error = nil, want foreign-key error")
	}
}
