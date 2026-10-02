package storage

import (
	"context"
	"testing"

	"github.com/umam-src/klip/internal/domain"
)

func TestSasaranRepositoryLifecycle(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRepository(db)
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatal(err)
	}

	sasaran := domain.Sasaran{ID: "sasaran-1", RuangID: "ruang-1", Title: "Rilis awal"}
	if err := repo.CreateSasaran(ctx, sasaran); err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetSasaran(ctx, sasaran.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.StatusDraft || got.Title != sasaran.Title || got.RuangID != sasaran.RuangID {
		t.Fatalf("sasaran tidak sesuai: %+v", got)
	}

	items, err := repo.ListSasaranByRuang(ctx, sasaran.RuangID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != sasaran.ID {
		t.Fatalf("daftar sasaran tidak sesuai: %+v", items)
	}
}

func TestSasaranRepositoryRejectsMissingSpace(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRepository(db)
	err = repo.CreateSasaran(ctx, domain.Sasaran{ID: "sasaran-1", Title: "Tanpa ruang"})
	if err == nil || !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}
