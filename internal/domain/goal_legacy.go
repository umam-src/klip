package domain

// GoalFromSasaran adapts the legacy Sasaran representation to the Goal
// contract without changing storage. ID and workspace identity are preserved.
//
// Sasaran used the generic Status lifecycle, so only completed has a direct
// Goal equivalent. Other non-terminal legacy states remain active; cancelled
// and failed records are intentionally not treated as completed.
func GoalFromSasaran(s Sasaran) Goal {
	status := GoalStatusActive
	if s.Status == StatusCompleted {
		status = GoalStatusCompleted
	}

	return Goal{
		ID:        s.ID,
		RuangID:   s.RuangID,
		Title:     s.Title,
		Status:    status,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}
