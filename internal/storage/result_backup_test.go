package storage

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestBackupResults(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "proyek-1"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "proyek-1", "hasil.txt"), []byte("halo Klip"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "proyek-1", "hasil.txt"), filepath.Join(root, "tautan.txt")); err != nil {
		t.Logf("symlink tidak tersedia di lingkungan ini, pemeriksaan tautan dilewati: %v", err)
	}

	archivePath := filepath.Join(t.TempDir(), "backup", "hasil.tar.gz")
	if err := BackupResults(root, archivePath); err != nil {
		t.Fatal(err)
	}

	file, err := os.Open(archivePath)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	gz, err := gzip.NewReader(file)
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)

	header, err := tr.Next()
	if err != nil {
		t.Fatal(err)
	}
	if header.Name != "proyek-1/hasil.txt" {
		t.Fatalf("header.Name = %q, want hasil path", header.Name)
	}
	data, err := io.ReadAll(tr)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "halo Klip" {
		t.Fatalf("content = %q, want %q", data, "halo Klip")
	}
	if _, err := tr.Next(); err != io.EOF {
		t.Fatalf("expected one archive entry, got err=%v", err)
	}
}

func TestBackupResultsRejectsExistingDestination(t *testing.T) {
	root := t.TempDir()
	destination := filepath.Join(t.TempDir(), "hasil.tar.gz")
	if err := os.WriteFile(destination, []byte("existing"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := BackupResults(root, destination); err == nil {
		t.Fatal("expected existing destination to be rejected")
	}
}
