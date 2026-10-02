package domain

import "testing"

func TestStatusCanTransitionTo(t *testing.T) {
	tests := []struct {
		name string
		from Status
		to   Status
		want bool
	}{
		{name: "draft to ready", from: StatusDraft, to: StatusReady, want: true},
		{name: "ready to running", from: StatusReady, to: StatusRunning, want: true},
		{name: "running to waiting", from: StatusRunning, to: StatusWaiting, want: true},
		{name: "waiting to running", from: StatusWaiting, to: StatusRunning, want: true},
		{name: "blocked to ready", from: StatusBlocked, to: StatusReady, want: true},
		{name: "running to completed", from: StatusRunning, to: StatusCompleted, want: true},
		{name: "running to failed", from: StatusRunning, to: StatusFailed, want: true},
		{name: "running to cancelled", from: StatusRunning, to: StatusCancelled, want: true},
		{name: "completed stays completed", from: StatusCompleted, to: StatusCompleted, want: true},
		{name: "failed stays failed", from: StatusFailed, to: StatusFailed, want: true},
		{name: "cancelled stays cancelled", from: StatusCancelled, to: StatusCancelled, want: true},
		{name: "draft cannot complete directly", from: StatusDraft, to: StatusCompleted, want: false},
		{name: "completed cannot restart", from: StatusCompleted, to: StatusRunning, want: false},
		{name: "unknown status rejected", from: Status("unknown"), to: StatusReady, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransitionTo(tt.to); got != tt.want {
				t.Fatalf("CanTransitionTo(%q) = %v, want %v", tt.to, got, tt.want)
			}
		})
	}
}

func TestStatusCanTransitionToSameStatus(t *testing.T) {
	for _, status := range []Status{
		StatusDraft,
		StatusReady,
		StatusRunning,
		StatusWaiting,
		StatusBlocked,
		StatusCompleted,
		StatusFailed,
		StatusCancelled,
	} {
		if !status.CanTransitionTo(status) {
			t.Fatalf("status %q should allow idempotent transition", status)
		}
	}
}
