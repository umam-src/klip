package domain

import "time"

// Proyek represents operational work that contributes to one Goal.
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
