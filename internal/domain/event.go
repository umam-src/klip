package domain

import "time"

const (
	EventExecutionStarted   = "execution.started"
	EventExecutionCompleted = "execution.completed"
	EventExecutionFailed    = "execution.failed"
	EventExecutionCancelled = "execution.cancelled"
)

type Event struct {
	ID          ID         `json:"id"`
	PekerjaanID ID         `json:"pekerjaan_id"`
	TugasID     *ID        `json:"tugas_id,omitempty"`
	SesiID      *ID        `json:"sesi_id,omitempty"`
	RunID       *ID        `json:"run_id,omitempty"`
	AgenID      *ID        `json:"agen_id,omitempty"`
	Type        string     `json:"type"`
	Message     string     `json:"message,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

func EventTypeForStatus(status Status) string {
	switch status {
	case StatusCompleted:
		return EventExecutionCompleted
	case StatusCancelled:
		return EventExecutionCancelled
	case StatusFailed:
		return EventExecutionFailed
	default:
		return ""
	}
}
