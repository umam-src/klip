package domain

import "time"

type ApprovalStatus string

const (
	ApprovalPending  ApprovalStatus = "pending"
	ApprovalApproved ApprovalStatus = "approved"
	ApprovalRejected ApprovalStatus = "rejected"
)

type Approval struct {
	ID        ID             `json:"id"`
	RuangID   ID             `json:"ruang_id"`
	ProyekID  ID             `json:"proyek_id"`
	TugasID   *ID            `json:"tugas_id,omitempty"`
	Status    ApprovalStatus `json:"status"`
	Reason    string         `json:"reason,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DecidedAt *time.Time     `json:"decided_at,omitempty"`
}
