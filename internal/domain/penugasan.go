package domain

import "time"

// Penugasan menghubungkan satu Tugas dengan satu Agen.
// Satu Tugas dapat memiliki beberapa Penugasan dan satu Agen dapat menerima
// beberapa Penugasan.
type Penugasan struct {
	ID        ID        `json:"id"`
	TugasID   ID        `json:"tugas_id"`
	AgenID    ID        `json:"agen_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
