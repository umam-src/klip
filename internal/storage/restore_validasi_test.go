package storage

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/umam-src/klip/internal/domain"
)

// siapkanCadanganValid membuat database v1 berisi satu Ruang Kerja lalu mencadangkannya.
func siapkanCadanganValid(t *testing.T, dir string) string {
	t.Helper()
	ctx := context.Background()
	db, err := Open(ctx, filepath.Join(dir, "sumber.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := NewRepository(db).CreateRuang(ctx, domain.Ruang{ID: "ruang-1", Name: "Ruang Kerja Cadangan"}); err != nil {
		t.Fatal(err)
	}
	cadangan := filepath.Join(dir, "cadangan.db")
	if err := BackupDatabase(ctx, db, cadangan); err != nil {
		t.Fatal(err)
	}
	return cadangan
}

func TestRestoreDatabaseMenolakSumberTidakValid(t *testing.T) {
	versiLain, err := os.CreateTemp(t.TempDir(), "versi-*.db")
	if err != nil {
		t.Fatal(err)
	}
	versiLain.Close()
	raw, err := sql.Open("sqlite", versiLain.Name())
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []string{
		`CREATE TABLE schema_migrations (version INTEGER PRIMARY KEY)`,
		`INSERT INTO schema_migrations(version) VALUES (9)`,
	} {
		if _, err := raw.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	raw.Close()

	cases := []struct {
		name   string
		isi    []byte
		sumber string
	}{
		{name: "bukan SQLite", isi: []byte("ini bukan database sqlite, hanya teks biasa yang cukup panjang untuk dibaca")},
		{name: "berkas kosong", isi: []byte{}},
		{name: "versi skema lain", sumber: versiLain.Name()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			sumber := tc.sumber
			if sumber == "" {
				sumber = filepath.Join(dir, "sumber.db")
				if err := os.WriteFile(sumber, tc.isi, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			tujuan := filepath.Join(dir, "klip.db")
			asli := []byte("database tujuan yang sehat")
			if err := os.WriteFile(tujuan, asli, 0o600); err != nil {
				t.Fatal(err)
			}

			if err := RestoreDatabase(sumber, tujuan); err == nil {
				t.Fatal("RestoreDatabase() menerima sumber yang tidak valid")
			}
			got, err := os.ReadFile(tujuan)
			if err != nil || string(got) != string(asli) {
				t.Fatalf("database tujuan berubah setelah restore gagal: %q, %v", got, err)
			}
			sisa, _ := filepath.Glob(filepath.Join(dir, ".klip-restore-*"))
			if len(sisa) != 0 {
				t.Fatalf("file sementara tertinggal: %v", sisa)
			}
		})
	}
}

func TestRestoreDatabaseMembersihkanWALLama(t *testing.T) {
	dir := t.TempDir()
	cadangan := siapkanCadanganValid(t, dir)

	tujuan := filepath.Join(dir, "klip.db")
	for _, berkas := range []string{tujuan, tujuan + "-wal", tujuan + "-shm"} {
		if err := os.WriteFile(berkas, []byte("sisa database lama"), 0o600); err != nil {
			t.Fatal(err)
		}
	}

	if err := RestoreDatabase(cadangan, tujuan); err != nil {
		t.Fatalf("RestoreDatabase() error = %v", err)
	}
	for _, suffix := range []string{"-wal", "-shm"} {
		if _, err := os.Stat(tujuan + suffix); !os.IsNotExist(err) {
			t.Fatalf("berkas %s lama masih ada: %v", suffix, err)
		}
	}

	ctx := context.Background()
	db, err := Open(ctx, tujuan)
	if err != nil {
		t.Fatalf("Open() hasil restore error = %v", err)
	}
	defer db.Close()
	ruang, err := NewRepository(db).GetRuang(ctx, "ruang-1")
	if err != nil || ruang.Name != "Ruang Kerja Cadangan" {
		t.Fatalf("data hasil restore = %+v, %v", ruang, err)
	}
}
