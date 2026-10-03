package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/umam-src/klip/internal/domain"
)

// Regresi: database v1 yang sudah ada harus bisa dibuka ulang (restart aplikasi)
// tanpa kehilangan data. Tes lain memakai :memory:, sehingga bug ini tidak terdeteksi.
func TestOpenUlangDatabaseV1MempertahankanData(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "klip.db")

	db, err := Open(ctx, path)
	if err != nil {
		t.Fatalf("Open() pertama error = %v", err)
	}
	if err := NewRepository(db).CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang Kerja Uji"}); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}

	db, err = Open(ctx, path)
	if err != nil {
		t.Fatalf("Open() ulang atas database v1 yang valid error = %v", err)
	}
	defer db.Close()

	ruang, err := NewRepository(db).GetRuang(ctx, "ruang-1")
	if err != nil {
		t.Fatalf("GetRuang() setelah dibuka ulang error = %v", err)
	}
	if ruang.Name != "Ruang Kerja Uji" {
		t.Fatalf("nama ruang = %q, want %q", ruang.Name, "Ruang Kerja Uji")
	}
}
