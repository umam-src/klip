package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/umam-src/klip/internal/domain"
)

func TestPekerjaanRejectsCrossRuangSasaran(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, "file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	repo := NewRepository(db)
	for _, ruang := range []domain.Ruang{
		{ID: "ruang-1", Name: "Ruang 1"},
		{ID: "ruang-2", Name: "Ruang 2"},
	} {
		if err := repo.CreateRuang(ctx, ruang); err != nil {
			t.Fatal(err)
		}
	}

	sasaran := domain.Sasaran{ID: "sasaran-2", RuangID: "ruang-2", Title: "Sasaran ruang 2"}
	if err := repo.CreateSasaran(ctx, sasaran); err != nil {
		t.Fatal(err)
	}

	err = repo.CreatePekerjaan(ctx, domain.Pekerjaan{
		ID:        "pekerjaan-1",
		RuangID:   "ruang-1",
		SasaranID: &sasaran.ID,
		Title:     "Pekerjaan lintas ruang",
		Status:    domain.StatusDraft,
	})
	if err == nil || !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected cross-ruang sasaran to be rejected, got %v", err)
	}
}
