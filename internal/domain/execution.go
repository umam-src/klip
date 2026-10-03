package domain

import "time"

// Eksekusi records one concrete process execution by an agent.
type Eksekusi struct {
	ID         ID         `json:"id"`
	RuangID    ID         `json:"ruang_id"`
	ProyekID   ID         `json:"proyek_id"`
	TugasID    *ID        `json:"tugas_id,omitempty"`
	AgenID     ID         `json:"agen_id"`
	Status     Status     `json:"status"`
	Program    string     `json:"program"`
	Arguments  []string   `json:"arguments"`
	ExitCode   *int       `json:"exit_code,omitempty"`
	Stdout     string     `json:"stdout"`
	Stderr     string     `json:"stderr"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}

// Hasil records an output or evidence produced by work.
type Hasil struct {
	ID         ID        `json:"id"`
	RuangID    ID        `json:"ruang_id"`
	EksekusiID *ID       `json:"eksekusi_id,omitempty"`
	ProyekID   ID        `json:"proyek_id"`
	TugasID    *ID       `json:"tugas_id,omitempty"`
	Kind       string    `json:"kind"`
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	CreatedAt  time.Time `json:"created_at"`
}
