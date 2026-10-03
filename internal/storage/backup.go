package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// BackupDatabase membuat salinan SQLite yang konsisten ke destination.
// VACUUM INTO digunakan agar WAL dan transaksi aktif ikut dipadatkan ke satu file.
func BackupDatabase(ctx context.Context, db *sql.DB, destination string) error {
	if db == nil {
		return errors.New("backup: database wajib diisi")
	}
	destination, err := prepareBackupPath(destination)
	if err != nil {
		return err
	}
	if _, err := os.Stat(destination); err == nil {
		return fmt.Errorf("backup: tujuan sudah ada: %w", os.ErrExist)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("backup: periksa tujuan: %w", err)
	}

	if _, err := db.ExecContext(ctx, "VACUUM INTO ?", destination); err != nil {
		return fmt.Errorf("backup: buat salinan: %w", err)
	}
	return nil
}

// RestoreDatabase mengganti destination dengan salinan SQLite source secara atomik.
// Pemanggil harus memastikan destination tidak sedang dibuka oleh koneksi SQLite.
func RestoreDatabase(source, destination string) error {
	if strings.TrimSpace(source) == "" || strings.TrimSpace(destination) == "" {
		return errors.New("restore: sumber dan tujuan wajib diisi")
	}
	if filepath.Clean(source) == filepath.Clean(destination) {
		return errors.New("restore: sumber dan tujuan harus berbeda")
	}

	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("restore: buka sumber: %w", err)
	}
	defer input.Close()

	info, err := input.Stat()
	if err != nil {
		return fmt.Errorf("restore: baca sumber: %w", err)
	}
	if !info.Mode().IsRegular() {
		return errors.New("restore: sumber bukan file biasa")
	}

	dir := filepath.Dir(destination)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("restore: buat direktori: %w", err)
	}
	temp, err := os.CreateTemp(dir, ".klip-restore-*")
	if err != nil {
		return fmt.Errorf("restore: buat file sementara: %w", err)
	}
	tempName := temp.Name()
	defer os.Remove(tempName)

	if _, err := io.Copy(temp, input); err != nil {
		_ = temp.Close()
		return fmt.Errorf("restore: salin database: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("restore: sinkronkan database: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("restore: tutup file sementara: %w", err)
	}
	if err := validateRestoreSource(tempName); err != nil {
		return err
	}
	if err := os.Rename(tempName, destination); err != nil {
		return fmt.Errorf("restore: ganti database: %w", err)
	}
	// Sisa WAL/SHM milik database lama tidak boleh terbaca terhadap database hasil restore.
	for _, suffix := range []string{"-wal", "-shm"} {
		if err := os.Remove(destination + suffix); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("restore: bersihkan berkas %s: %w", suffix, err)
		}
	}
	return nil
}

// validateRestoreSource memeriksa salinan sementara sebelum menimpa database tujuan:
// harus berkas SQLite yang utuh dan berversi skema Klip saat ini tanpa tabel legacy.
// Pemeriksaan dilakukan pada salinan, bukan pada berkas cadangan asli.
func validateRestoreSource(path string) error {
	ctx := context.Background()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("restore: buka sumber untuk diperiksa: %w", err)
	}
	defer func() {
		_ = db.Close()
		_ = os.Remove(path + "-wal")
		_ = os.Remove(path + "-shm")
	}()
	db.SetMaxOpenConns(1)

	var hasil string
	if err := db.QueryRowContext(ctx, `PRAGMA quick_check`).Scan(&hasil); err != nil || hasil != "ok" {
		return errors.New("restore: sumber bukan database SQLite yang utuh")
	}
	version, err := databaseSchemaVersion(ctx, db)
	if err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	if version != schemaVersion {
		return fmt.Errorf("restore: sumber bukan database Klip skema v%d (versi terbaca %d)", schemaVersion, version)
	}
	if err := rejectLegacySchema(ctx, db); err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	return nil
}

func prepareBackupPath(destination string) (string, error) {
	if strings.TrimSpace(destination) == "" {
		return "", errors.New("backup: tujuan wajib diisi")
	}
	clean := filepath.Clean(destination)
	if clean == "." || strings.HasSuffix(clean, string(filepath.Separator)) {
		return "", errors.New("backup: tujuan tidak valid")
	}
	if err := os.MkdirAll(filepath.Dir(clean), 0o755); err != nil {
		return "", fmt.Errorf("backup: buat direktori: %w", err)
	}
	return clean, nil
}
