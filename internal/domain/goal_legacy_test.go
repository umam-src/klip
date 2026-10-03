package domain

import (
	"testing"
	"time"
)

func TestGoalFromSasaranPreservesIdentityAndWorkspace(t *testing.T) {
	now := time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)
	sasaran := Sasaran{
		ID:        "goal-1",
		RuangID:   "ruang-1",
		Title:     "Tujuan lama",
		Status:    StatusReady,
		CreatedAt: now,
		UpdatedAt: now,
	}

	goal := GoalFromSasaran(sasaran)
	if goal.ID != sasaran.ID || goal.RuangID != sasaran.RuangID {
		t.Fatalf("identity mapping changed: goal=%+v sasaran=%+v", goal, sasaran)
	}
	if goal.Title != sasaran.Title {
		t.Fatalf("title = %q, want %q", goal.Title, sasaran.Title)
	}
	if goal.Status != GoalStatusActive {
		t.Fatalf("status = %q, want %q", goal.Status, GoalStatusActive)
	}
	if !goal.CreatedAt.Equal(now) || !goal.UpdatedAt.Equal(now) {
		t.Fatalf("timestamps were not preserved: goal=%+v", goal)
	}
}

func TestGoalFromSasaranCompleted(t *testing.T) {
	goal := GoalFromSasaran(Sasaran{
		ID:     "goal-1",
		RuangID: "ruang-1",
		Title:  "Selesai",
		Status: StatusCompleted,
	})

	if goal.Status != GoalStatusCompleted {
		t.Fatalf("status = %q, want %q", goal.Status, GoalStatusCompleted)
	}
}

func TestGoalFromSasaranDoesNotInventHierarchyOrDescription(t *testing.T) {
	goal := GoalFromSasaran(Sasaran{
		ID:      "goal-1",
		RuangID: "ruang-1",
		Title:   "Legacy",
		Status:  StatusDraft,
	})

	if goal.ParentGoalID != nil {
		t.Fatal("legacy Sasaran must not invent a parent Goal")
	}
	if goal.Description != "" {
		t.Fatal("legacy Sasaran must not invent a description")
	}
}
