package domain

import (
	"path"
	"strings"
	"time"
)

// HasilKerja adalah keluaran yang dihasilkan dari pekerjaan operasional Klip.
// Path selalu berupa referensi relatif di dalam penyimpanan lokal aplikasi.
type HasilKerja struct {
	ID          ID        `json:"id"`
	RuangID     ID        `json:"ruang_id"`
	ExecutionID *ID       `json:"execution_id,omitempty"`
	ProyekID    ID        `json:"proyek_id"`
	TugasID     *ID       `json:"tugas_id,omitempty"`
	Kind        string    `json:"kind"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	CreatedAt   time.Time `json:"created_at"`
}

func (h HasilKerja) Valid() bool {
	if strings.TrimSpace(string(h.ID)) == "" ||
		strings.TrimSpace(string(h.RuangID)) == "" ||
		strings.TrimSpace(string(h.ProyekID)) == "" ||
		strings.TrimSpace(h.Kind) == "" ||
		strings.TrimSpace(h.Name) == "" ||
		strings.TrimSpace(h.Path) == "" {
		return false
	}
	if strings.ContainsRune(h.Name, '\x00') || strings.ContainsAny(h.Name, `/\\`) {
		return false
	}
	if strings.ContainsRune(h.Path, '\x00') || strings.ContainsRune(h.Path, '\\') || path.IsAbs(h.Path) {
		return false
	}
	clean := path.Clean(h.Path)
	return clean == h.Path && clean != "." && clean != ".." && !strings.HasPrefix(clean, "../")
}
