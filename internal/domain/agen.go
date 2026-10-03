package domain

import (
	"errors"
	"strings"
)

var ErrInvalidAgen = errors.New("agen tidak valid")

// Validate checks invariants that can be enforced without reading storage.
// Rules involving the Ruang or parent hierarchy remain in the storage layer.
func (a Agen) Validate() error {
	if strings.TrimSpace(string(a.ID)) == "" {
		return errors.Join(ErrInvalidAgen, errors.New("id wajib diisi"))
	}
	if strings.TrimSpace(string(a.RuangID)) == "" {
		return errors.Join(ErrInvalidAgen, errors.New("ruang wajib diisi"))
	}
	if strings.TrimSpace(a.Name) == "" {
		return errors.Join(ErrInvalidAgen, errors.New("nama wajib diisi"))
	}
	if strings.TrimSpace(a.Role) == "" {
		return errors.Join(ErrInvalidAgen, errors.New("peran wajib diisi"))
	}
	if a.Status != "" && !a.Status.IsKnown() {
		return errors.Join(ErrInvalidAgen, errors.New("status agen tidak dikenal"))
	}
	if a.ParentID != nil && strings.TrimSpace(string(*a.ParentID)) == string(a.ID) {
		return errors.Join(ErrInvalidAgen, errors.New("agen tidak boleh menjadi induknya sendiri"))
	}
	return nil
}
