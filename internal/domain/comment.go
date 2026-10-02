package domain

import "time"

// Komentar menyimpan percakapan ringan yang terkait dengan pekerjaan atau tugas.
type Komentar struct {
	ID          ID        `json:"id"`
	PekerjaanID ID        `json:"pekerjaan_id"`
	TugasID     *ID       `json:"tugas_id,omitempty"`
	ParentID    *ID       `json:"parent_id,omitempty"`
	Body        string    `json:"body"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
