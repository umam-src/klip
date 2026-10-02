package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const maxResultSize int64 = 16 << 20

// ResultStore menyimpan berkas hasil di bawah direktori hasil Klip.
// Path yang diberikan pemanggil selalu relatif terhadap root dan tidak boleh keluar.
type ResultStore struct {
	root string
}

func NewResultStore(dataDir string) (*ResultStore, error) {
	if strings.TrimSpace(dataDir) == "" {
		return nil, errors.New("hasil: direktori data wajib diisi")
	}
	root, err := filepath.Abs(filepath.Join(dataDir, "hasil"))
	if err != nil {
		return nil, fmt.Errorf("hasil: tentukan direktori: %w", err)
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("hasil: buat direktori: %w", err)
	}
	if err := ensureNoSymlinkPath(root, root); err != nil {
		return nil, err
	}
	return &ResultStore{root: root}, nil
}

func (s *ResultStore) Put(relPath string, r io.Reader) error {
	path, err := s.resolve(relPath)
	if err != nil {
		return err
	}
	parent := filepath.Dir(path)
	if err := ensureNoSymlinkPath(s.root, parent); err != nil {
		return err
	}
	if err := os.MkdirAll(parent, 0o700); err != nil {
		return fmt.Errorf("hasil: buat direktori: %w", err)
	}
	if err := ensureNoSymlinkPath(s.root, parent); err != nil {
		return err
	}

	temp, err := os.CreateTemp(parent, ".klip-result-*")
	if err != nil {
		return fmt.Errorf("hasil: buat berkas sementara: %w", err)
	}
	tempName := temp.Name()
	defer os.Remove(tempName)

	if _, err := io.Copy(temp, io.LimitReader(r, maxResultSize+1)); err != nil {
		_ = temp.Close()
		return fmt.Errorf("hasil: tulis berkas: %w", err)
	}
	info, err := temp.Stat()
	if err != nil {
		_ = temp.Close()
		return fmt.Errorf("hasil: periksa berkas: %w", err)
	}
	if info.Size() > maxResultSize {
		_ = temp.Close()
		return fmt.Errorf("hasil: ukuran berkas melebihi %d byte: %w", maxResultSize, ErrInvalid)
	}
	if err := temp.Sync(); err != nil {
		_ = temp.Close()
		return fmt.Errorf("hasil: sinkronkan berkas: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("hasil: tutup berkas: %w", err)
	}
	if err := os.Rename(tempName, path); err != nil {
		return fmt.Errorf("hasil: simpan berkas: %w", err)
	}
	return nil
}

func (s *ResultStore) Read(relPath string) ([]byte, error) {
	path, err := s.resolve(relPath)
	if err != nil {
		return nil, err
	}
	if err := ensureNoSymlinkPath(s.root, filepath.Dir(path)); err != nil {
		return nil, err
	}
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("hasil: periksa berkas: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return nil, fmt.Errorf("hasil: berkas tidak valid: %w", ErrInvalid)
	}
	if info.Size() > maxResultSize {
		return nil, fmt.Errorf("hasil: ukuran berkas tidak valid: %w", ErrInvalid)
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("hasil: baca berkas: %w", err)
	}
	return data, nil
}

func (s *ResultStore) resolve(relPath string) (string, error) {
	relPath = filepath.FromSlash(strings.TrimSpace(relPath))
	if relPath == "" || filepath.IsAbs(relPath) || relPath == "." || relPath == ".." || strings.ContainsRune(relPath, 0) {
		return "", fmt.Errorf("hasil: path tidak valid: %w", ErrInvalid)
	}
	clean := filepath.Clean(relPath)
	if clean != relPath || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("hasil: path keluar dari direktori hasil: %w", ErrInvalid)
	}
	path := filepath.Join(s.root, clean)
	if err := ensureWithin(s.root, path); err != nil {
		return "", err
	}
	return path, nil
}

func ensureWithin(root, candidate string) error {
	rel, err := filepath.Rel(root, candidate)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("hasil: path keluar dari direktori hasil: %w", ErrInvalid)
	}
	return nil
}

func ensureNoSymlinkPath(root, candidate string) error {
	if err := ensureWithin(root, candidate); err != nil {
		return err
	}
	rel, err := filepath.Rel(root, candidate)
	if err != nil {
		return fmt.Errorf("hasil: periksa direktori: %w", ErrInvalid)
	}
	current := root
	if rel == "." {
		return ensureDirectory(current)
	}
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("hasil: periksa direktori: %w", err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("hasil: direktori tidak boleh berupa symlink: %w", ErrInvalid)
		}
		if !info.IsDir() {
			return fmt.Errorf("hasil: path induk bukan direktori: %w", ErrInvalid)
		}
	}
	return nil
}

func ensureDirectory(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return fmt.Errorf("hasil: periksa direktori: %w", err)
	}
	if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return fmt.Errorf("hasil: root hasil tidak valid: %w", ErrInvalid)
	}
	return nil
}
