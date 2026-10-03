package domain

import "testing"

func TestGoalStatusIsKnown(t *testing.T) {
	tests := []struct {
		status GoalStatus
		want   bool
	}{
		{status: GoalStatusActive, want: true},
		{status: GoalStatusCompleted, want: true},
		{status: GoalStatusArchived, want: true},
		{status: GoalStatus("unknown"), want: false},
	}

	for _, tt := range tests {
		if got := tt.status.IsKnown(); got != tt.want {
			t.Fatalf("GoalStatus(%q).IsKnown() = %v, want %v", tt.status, got, tt.want)
		}
	}
}

func TestGoalStatusCanTransitionTo(t *testing.T) {
	tests := []struct {
		name string
		from GoalStatus
		to   GoalStatus
		want bool
	}{
		{name: "active stays active", from: GoalStatusActive, to: GoalStatusActive, want: true},
		{name: "active to completed", from: GoalStatusActive, to: GoalStatusCompleted, want: true},
		{name: "active to archived", from: GoalStatusActive, to: GoalStatusArchived, want: true},
		{name: "completed stays completed", from: GoalStatusCompleted, to: GoalStatusCompleted, want: true},
		{name: "completed to archived", from: GoalStatusCompleted, to: GoalStatusArchived, want: true},
		{name: "archived stays archived", from: GoalStatusArchived, to: GoalStatusArchived, want: true},
		{name: "completed cannot become active", from: GoalStatusCompleted, to: GoalStatusActive, want: false},
		{name: "archived cannot become active", from: GoalStatusArchived, to: GoalStatusActive, want: false},
		{name: "archived cannot become completed", from: GoalStatusArchived, to: GoalStatusCompleted, want: false},
		{name: "unknown status rejected", from: GoalStatus("unknown"), to: GoalStatusActive, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.want {
				t.Fatalf("CanTransitionTo(%q) = %v, want %v", tt.to, got, tt.want)
			}
		})
	}
}

func TestGoalSupportsRootAndChildReferences(t *testing.T) {
	parent := ID("goal-parent")

	root := Goal{
		ID:      ID("goal-root"),
		RuangID: ID("ruang-1"),
		Status:  GoalStatusActive,
	}
	if root.ParentGoalID != nil {
		t.Fatal("root Goal should not require a parent")
	}

	child := Goal{
		ID:           ID("goal-child"),
		RuangID:      ID("ruang-1"),
		ParentGoalID: &parent,
		Status:       GoalStatusActive,
	}
	if child.ParentGoalID == nil || *child.ParentGoalID != parent {
		t.Fatalf("child Goal parent = %v, want %q", child.ParentGoalID, parent)
	}
}
