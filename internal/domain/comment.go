package domain

import "time"

// Komentar menyimpan percakapan ringan yang terkait dengan Proyek atau Tugas.
// Komentar tanpa TugasID berada pada tingkat Proyek.
type Komentar struct {
	ID        ID        `json:"id"`
	RuangID   ID        `json:"ruang_id"`
	ProyekID  ID        `json:"proyek_id"`
	TugasID   *ID       `json:"tugas_id,omitempty"`
	ParentID  *ID       `json:"parent_id,omitempty"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
