package domain

import "time"

type ID string

type Status string

const (
	StatusDraft     Status = "draft"
	StatusReady     Status = "ready"
	StatusRunning   Status = "running"
	StatusWaiting   Status = "waiting"
	StatusBlocked   Status = "blocked"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusCancelled Status = "cancelled"
)

type Ruang struct {
	ID        ID        `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Agen struct {
	ID          ID        `json:"id"`
	RuangID     ID        `json:"ruang_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	ProviderID  string    `json:"provider_id,omitempty"`
	ModelID     string    `json:"model_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Sasaran struct {
	ID        ID        `json:"id"`
	RuangID   ID        `json:"ruang_id"`
	Title     string    `json:"title"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Pekerjaan struct {
	ID        ID        `json:"id"`
	RuangID   ID        `json:"ruang_id"`
	SasaranID *ID       `json:"sasaran_id,omitempty"`
	Title     string    `json:"title"`
	Status    Status    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tugas struct {
	ID          ID        `json:"id"`
	PekerjaanID ID        `json:"pekerjaan_id"`
	ParentID    *ID       `json:"parent_id,omitempty"`
	Title       string    `json:"title"`
	Status      Status    `json:"status"`
	Position    int       `json:"position"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Sesi struct {
	ID          ID         `json:"id"`
	PekerjaanID ID         `json:"pekerjaan_id"`
	AgenID      ID         `json:"agen_id"`
	Status      Status     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
}

type Hasil struct {
	ID          ID        `json:"id"`
	PekerjaanID ID        `json:"pekerjaan_id"`
	TugasID     *ID       `json:"tugas_id,omitempty"`
	Kind        string    `json:"kind"`
	Name        string    `json:"name"`
	Path        string    `json:"path"`
	CreatedAt   time.Time `json:"created_at"`
}

type Run struct {
	ID          ID         `json:"id"`
	PekerjaanID ID         `json:"pekerjaan_id"`
	TugasID     *ID        `json:"tugas_id,omitempty"`
	AgenID      ID         `json:"agen_id"`
	Status      Status     `json:"status"`
	Program     string     `json:"program"`
	Arguments   []string   `json:"arguments"`
	ExitCode    *int       `json:"exit_code,omitempty"`
	Stdout      string     `json:"stdout"`
	Stderr      string     `json:"stderr"`
	StartedAt   time.Time  `json:"started_at"`
	FinishedAt  *time.Time `json:"finished_at,omitempty"`
}
