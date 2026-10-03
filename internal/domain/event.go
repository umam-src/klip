package domain

import "time"

const (
	EventExecutionStarted   = "execution.started"
	EventExecutionCompleted = "execution.completed"
	EventExecutionFailed    = "execution.failed"
	EventExecutionCancelled = "execution.cancelled"
	EventApprovalCreated    = "approval.created"
	EventApprovalDecided    = "approval.decided"
	EventScheduleCreated    = "schedule.created"
)

type Event struct {
	ID         ID        `json:"id"`
	RuangID    ID        `json:"ruang_id"`
	ExecutionID *ID      `json:"execution_id,omitempty"`
	ProyekID   ID        `json:"proyek_id"`
	TugasID    *ID       `json:"tugas_id,omitempty"`
	AgenID     *ID       `json:"agen_id,omitempty"`
	Type       string    `json:"type"`
	Message    string    `json:"message,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
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
