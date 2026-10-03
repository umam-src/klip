package app

import (
	"path"
	"strings"

	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

func validateHasil(hasil domain.HasilKerja) error {
	if strings.TrimSpace(string(hasil.ID)) == "" ||
		strings.TrimSpace(string(hasil.ProyekID)) == "" ||
		strings.TrimSpace(hasil.Kind) == "" ||
		strings.TrimSpace(hasil.Name) == "" ||
		strings.TrimSpace(hasil.Path) == "" {
		return storage.ErrInvalid
	}
	if strings.ContainsRune(hasil.Name, '\x00') || strings.ContainsAny(hasil.Name, `/\\`) {
		return storage.ErrInvalid
	}
	if strings.ContainsRune(hasil.Path, '\x00') || strings.ContainsRune(hasil.Path, '\\') || path.IsAbs(hasil.Path) {
		return storage.ErrInvalid
	}
	clean := path.Clean(hasil.Path)
	if clean != hasil.Path || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return storage.ErrInvalid
	}
	return nil
}
