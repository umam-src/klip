package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/umam-src/klip/internal/domain"
)

func TestUpdateAgenWithParent(t *testing.T) {
	ctx := context.Background()
	db, err := Open(ctx, ":memory:")
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	repo := NewRepository(db)
	for _, ruang := range []domain.Ruang{
		{ID: "ruang-1", Name: "Ruang"},
		{ID: "ruang-2", Name: "Ruang lain"},
	} {
		if err := repo.CreateRuang(ctx, ruang); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.CreateAgenWithParent(ctx, domain.Agen{ID: "agen-root", RuangID: "ruang-1", Name: "Root", Role: "Pemimpin"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgenWithParent(ctx, domain.Agen{ID: "agen-child", RuangID: "ruang-1", Name: "Child", Role: "Pelaksana"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgenWithParent(ctx, domain.Agen{ID: "agen-other", RuangID: "ruang-1", Name: "Other", Role: "Pelaksana"}); err != nil {
		t.Fatal(err)
	}
	if err := repo.CreateAgenWithParent(ctx, domain.Agen{ID: "agen-cross", RuangID: "ruang-2", Name: "Cross", Role: "Pelaksana"}); err != nil {
		t.Fatal(err)
	}

	root := domain.ID("agen-root")
	if err := repo.UpdateAgenWithParent(ctx, domain.Agen{ID: "agen-child", RuangID: "ruang-1", ParentID: &root, Name: "Child Baru", Role: "Operator", Description: "Deskripsi", ProviderID: "openai-compatible", ModelID: "model-1", Status: domain.AgenStatusInactive}); err != nil {
		t.Fatalf("UpdateAgenWithParent() error = %v", err)
	}
	updated, err := repo.GetAgenWithParent(ctx, "agen-child")
	if err != nil {
		t.Fatal(err)
	}
	if updated.ParentID == nil || *updated.ParentID != root || updated.Name != "Child Baru" || updated.Role != "Operator" || updated.Status != domain.AgenStatusInactive || updated.ModelID != "model-1" {
		t.Fatalf("updated agent = %+v", updated)
	}

	cross := domain.ID("agen-cross")
	if err := repo.UpdateAgenWithParent(ctx, domain.Agen{ID: "agen-child", RuangID: "ruang-1", ParentID: &cross, Name: "Child", Role: "Operator"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("cross-room parent error = %v, want ErrInvalid", err)
	}

	self := domain.ID("agen-child")
	if err := repo.UpdateAgenWithParent(ctx, domain.Agen{ID: "agen-child", RuangID: "ruang-1", ParentID: &self, Name: "Child", Role: "Operator"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("self parent error = %v, want ErrInvalid", err)
	}

	child := domain.ID("agen-child")
	if err := repo.UpdateAgenWithParent(ctx, domain.Agen{ID: "agen-root", RuangID: "ruang-1", ParentID: &child, Name: "Root", Role: "Pemimpin"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("cycle error = %v, want ErrInvalid", err)
	}

	missing := domain.ID("missing")
	if err := repo.UpdateAgenWithParent(ctx, domain.Agen{ID: "agen-child", RuangID: "ruang-1", ParentID: &missing, Name: "Child", Role: "Operator"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("missing parent error = %v, want ErrInvalid", err)
	}
}
