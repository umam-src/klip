package domain

import "time"

type ID string

type Status string

const (
	StatusDraft Status = "draft"
	StatusReady Status = "ready"
	StatusRunning Status = "running"
	StatusWaiting Status = "waiting"
	StatusBlocked Status = "blocked"
	StatusCompleted Status = "completed"
	StatusFailed Status = "failed"
	StatusCancelled Status = "cancelled"
)

func (s Status) CanTransitionTo(next Status) bool {
	if s == next {
		return isKnownStatus(s)
	}
	switch s {
	case StatusDraft:
		return next == StatusReady || next == StatusCancelled
	case StatusReady:
		return next == StatusRunning || next == StatusBlocked || next == StatusCancelled
	case StatusRunning:
		return next == StatusWaiting || next == StatusCompleted || next == StatusFailed || next == StatusCancelled
	case StatusWaiting:
		return next == StatusRunning || next == StatusBlocked || next == StatusCancelled
	case StatusBlocked:
		return next == StatusReady || next == StatusCancelled
	default:
		return false
	}
}

func isKnownStatus(status Status) bool {
	switch status {
	case StatusDraft, StatusReady, StatusRunning, StatusWaiting, StatusBlocked, StatusCompleted, StatusFailed, StatusCancelled:
		return true
	default:
		return false
	}
}

type AgenStatus string

const (
	AgenStatusActive AgenStatus = "active"
	AgenStatusInactive AgenStatus = "inactive"
)

func (s AgenStatus) IsKnown() bool {
	return s == AgenStatusActive || s == AgenStatusInactive
}

type Ruang struct {
	ID        ID        `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Agen struct {
	ID         ID         `json:"id"`
	RuangID    ID         `json:"ruang_id"`
	ParentID   *ID        `json:"parent_id,omitempty"`
	Name       string     `json:"name"`
	Role       string     `json:"role"`
	Description string    `json:"description,omitempty"`
	ProviderID string     `json:"provider_id,omitempty"`
	ModelID    string     `json:"model_id,omitempty"`
	Status     AgenStatus `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type Proyek struct {
	ID          ID        `json:"id"`
	RuangID     ID        `json:"ruang_id"`
	GoalID      ID        `json:"goal_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Penugasan struct {
	ID        ID        `json:"id"`
	TugasID   ID        `json:"tugas_id"`
	AgenID    ID        `json:"agen_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tugas struct {
	ID          ID        `json:"id"`
	RuangID     ID        `json:"ruang_id"`
	ProyekID    ID        `json:"proyek_id"`
	ParentID    *ID       `json:"parent_id,omitempty"`
	Title       string    `json:"title"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

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
	Stdout     string     `json:"stdout,omitempty"`
	Stderr     string     `json:"stderr,omitempty"`
	StartedAt  time.Time  `json:"started_at"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
}
