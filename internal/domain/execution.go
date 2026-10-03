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
