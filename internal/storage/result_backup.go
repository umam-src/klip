package storage

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupResults membuat arsip gzip berisi berkas hasil lokal.
// Tautan simbolik dan file non-reguler diabaikan agar arsip tidak mengikuti path keluar root.
func BackupResults(root, destination string) error {
	if strings.TrimSpace(root) == "" || strings.TrimSpace(destination) == "" {
		return errors.New("backup hasil: sumber dan tujuan wajib diisi")
	}
	root, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("backup hasil: tentukan sumber: %w", err)
	}
	info, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("backup hasil: buka sumber: %w", err)
	}
	if !info.IsDir() {
		return errors.New("backup hasil: sumber bukan direktori")
	}

	destination, err = filepath.Abs(destination)
	if err != nil {
		return fmt.Errorf("backup hasil: tentukan tujuan: %w", err)
	}
	if _, err := os.Lstat(destination); err == nil {
		return fmt.Errorf("backup hasil: tujuan sudah ada: %w", os.ErrExist)
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("backup hasil: periksa tujuan: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(destination), 0o700); err != nil {
		return fmt.Errorf("backup hasil: buat direktori tujuan: %w", err)
	}

	temp, err := os.CreateTemp(filepath.Dir(destination), ".klip-result-backup-*.tmp")
	if err != nil {
		return fmt.Errorf("backup hasil: buat file sementara: %w", err)
	}
	tempName := temp.Name()
	defer os.Remove(tempName)

	gz := gzip.NewWriter(temp)
	tarWriter := tar.NewWriter(gz)
	files, err := resultFiles(root)
	if err != nil {
		_ = tarWriter.Close()
		_ = gz.Close()
		_ = temp.Close()
		return err
	}
	for _, path := range files {
		if err := writeResultTar(tarWriter, root, path); err != nil {
			_ = tarWriter.Close()
			_ = gz.Close()
			_ = temp.Close()
			return err
		}
	}
	if err := tarWriter.Close(); err != nil {
		_ = gz.Close()
		_ = temp.Close()
		return fmt.Errorf("backup hasil: tutup arsip: %w", err)
	}
	if err := gz.Close(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("backup hasil: tutup gzip: %w", err)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("backup hasil: sinkronkan arsip: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("backup hasil: tutup file sementara: %w", err)
	}
	if err := os.Rename(tempName, destination); err != nil {
		return fmt.Errorf("backup hasil: simpan arsip: %w", err)
	}
	return nil
}

func resultFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == root {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("backup hasil: baca metadata %q: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("backup hasil: telusuri sumber: %w", err)
	}
	sort.Strings(files)
	return files, nil
}

func writeResultTar(tw *tar.Writer, root, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("backup hasil: baca %q: %w", path, err)
	}
	if !info.Mode().IsRegular() {
		return nil
	}
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("backup hasil: path tidak valid: %q", path)
	}
	input, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("backup hasil: buka %q: %w", path, err)
	}
	defer input.Close()

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return fmt.Errorf("backup hasil: metadata %q: %w", path, err)
	}
	header.Name = filepath.ToSlash(rel)
	header.ModTime = time.Unix(0, 0).UTC()
	if err := tw.WriteHeader(header); err != nil {
		return fmt.Errorf("backup hasil: tulis metadata %q: %w", path, err)
	}
	if _, err := io.Copy(tw, input); err != nil {
		return fmt.Errorf("backup hasil: salin %q: %w", path, err)
	}
	return nil
}
