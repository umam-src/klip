package domain

import "time"

// GoalStatus describes the lifecycle of a Goal.
type GoalStatus string

const (
	GoalStatusActive    GoalStatus = "active"
	GoalStatusCompleted GoalStatus = "completed"
	GoalStatusArchived  GoalStatus = "archived"
)

// IsKnown reports whether the status is part of the Goal lifecycle.
func (s GoalStatus) IsKnown() bool {
	switch s {
	case GoalStatusActive, GoalStatusCompleted, GoalStatusArchived:
		return true
	default:
		return false
	}
}

// CanTransitionTo reports whether a Goal can move to the target status.
func (s GoalStatus) CanTransitionTo(next GoalStatus) bool {
	if s == next {
		return s.IsKnown()
	}

	switch s {
	case GoalStatusActive:
		return next == GoalStatusCompleted || next == GoalStatusArchived
	case GoalStatusCompleted:
		return next == GoalStatusArchived
	default:
		return false
	}
}

// Goal represents a desired result or state within a workspace.
//
// Parent/child and workspace consistency are enforced by the storage/service
// layer because they require access to related records.
type Goal struct {
	ID           ID         `json:"id"`
	RuangID      ID         `json:"ruang_id"`
	ParentGoalID *ID        `json:"parent_goal_id,omitempty"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	Status       GoalStatus `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
