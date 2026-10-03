package storage

import (
	"context"
	"errors"
	"testing"

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
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Pemimpin", Role: "Pemimpin", Status: domain.AgenStatusActive}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Pekerjaan", Status: "open"}); err != nil {
		t.Fatalf("CreatePekerjaan() error = %v", err)
	}
	parent := domain.ID("tugas-1")
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: parent, PekerjaanID: "pekerjaan-1", Title: "Induk", Status: "open"}); err != nil {
		t.Fatalf("CreateTugas(parent) error = %v", err)
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-2", PekerjaanID: "pekerjaan-1", ParentID: &parent, Title: "Anak", Status: "open", Position: 2}); err != nil {
		t.Fatalf("CreateTugas(child) error = %v", err)
	}
	if err := repo.CreateHasil(ctx, domain.Hasil{ID: "hasil-1", PekerjaanID: "pekerjaan-1", TugasID: &parent, Kind: "file", Name: "hasil", Path: "hasil/hasil.txt"}); err != nil {
		t.Fatalf("CreateHasil() error = %v", err)
	}

	ruang, err := repo.GetRuang(ctx, "ruang-1")
	if err != nil || ruang.Name != "Ruang" {
		t.Fatalf("GetRuang() = %+v, error = %v", ruang, err)
	}

	agen, err := repo.GetAgen(ctx, "agen-1")
	if err != nil || agen.Role != "Pemimpin" || agen.Status != domain.AgenStatusActive {
		t.Fatalf("GetAgen() = %+v, error = %v", agen, err)
	}

	pekerjaan, err := repo.GetPekerjaan(ctx, "pekerjaan-1")
	if err != nil || pekerjaan.Title != "Pekerjaan" {
		t.Fatalf("GetPekerjaan() = %+v, error = %v", pekerjaan, err)
	}

	tugas, err := repo.GetTugas(ctx, "tugas-2")
	if err != nil || tugas.ParentID == nil || *tugas.ParentID != parent || tugas.Position != 2 {
		t.Fatalf("GetTugas() = %+v, error = %v", tugas, err)
	}

	hasil, err := repo.GetHasil(ctx, "hasil-1")
	if err != nil || hasil.TugasID == nil || *hasil.TugasID != parent || hasil.Path != "hasil/hasil.txt" {
		t.Fatalf("GetHasil() = %+v, error = %v", hasil, err)
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
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "legacy"}); err != nil {
		t.Fatalf("CreateRuang(legacy) error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "tanpa peran"}); err != nil {
		t.Fatalf("CreateAgen() legacy role normalization error = %v", err)
	}
	legacy, err := repo.GetAgen(ctx, "agen-1")
	if err != nil || legacy.Role != "Agen" || legacy.Status != domain.AgenStatusActive {
		t.Fatalf("legacy agent = %+v, error = %v, want role Agen and active status", legacy, err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-2", RuangID: "ruang-1", Name: "status salah", Role: "Pelaksana", Status: "unknown"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateAgen(status) error = %v, want ErrInvalid", err)
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "negatif", Position: -1}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateTugas() error = %v, want ErrInvalid", err)
	}
	if err := repo.CreateHasil(ctx, domain.Hasil{ID: "hasil-1", PekerjaanID: "pekerjaan-1", Kind: "file", Name: "hasil"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("CreateHasil() error = %v, want ErrInvalid", err)
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
	if err := repo.CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang"}); err != nil {
		t.Fatalf("CreateRuang() error = %v", err)
	}
	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-1", RuangID: "ruang-1", Name: "Agen", Role: "Pelaksana"}); err != nil {
		t.Fatalf("CreateAgen() error = %v", err)
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-1", RuangID: "ruang-1", Title: "Pekerjaan", Status: "open"}); err != nil {
		t.Fatalf("CreatePekerjaan() error = %v", err)
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-1", PekerjaanID: "pekerjaan-1", Title: "Tugas", Status: "open"}); err != nil {
		t.Fatalf("CreateTugas() error = %v", err)
	}
	if err := repo.CreateHasil(ctx, domain.Hasil{ID: "hasil-1", PekerjaanID: "pekerjaan-1", Kind: "file", Name: "hasil", Path: "hasil/hasil.txt"}); err != nil {
		t.Fatalf("CreateHasil() error = %v", err)
	}

	if err := repo.CreateAgen(ctx, domain.Agen{ID: "agen-invalid", RuangID: "ruang-missing", Name: "Agen", Role: "Pelaksana"}); err == nil {
		t.Fatal("CreateAgen() error = nil, want foreign key failure")
	}
	if err := repo.CreatePekerjaan(ctx, domain.Pekerjaan{ID: "pekerjaan-invalid", RuangID: "ruang-missing", Title: "Pekerjaan", Status: "open"}); err == nil {
		t.Fatal("CreatePekerjaan() error = nil, want foreign key failure")
	}
	if err := repo.CreateTugas(ctx, domain.Tugas{ID: "tugas-invalid", PekerjaanID: "pekerjaan-missing", Title: "Tugas", Status: "open"}); err == nil {
		t.Fatal("CreateTugas() error = nil, want foreign key failure")
	}
	if err := repo.CreateHasil(ctx, domain.Hasil{ID: "hasil-invalid", PekerjaanID: "pekerjaan-missing", Kind: "file", Name: "hasil"}); err == nil {
		t.Fatal("CreateHasil() error = nil, want foreign key failure")
	}
}
