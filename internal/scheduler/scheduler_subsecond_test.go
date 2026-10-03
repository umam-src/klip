package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

func TestSchedulerSubsecondIntervalAdvancesNextRun(t *testing.T) {
	start := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	interval := 500 * time.Millisecond
	repo := &fakeRepo{schedule: domain.Schedule{
		ID:         "s-subsecond",
		Name:       "uji sub-detik",
		RuangID:    "r1",
		ProyekID:   "p1",
		AgenID:     "a1",
		Program:    "true",
		Interval:   interval,
		NextRunAt:  start,
		Status:     domain.ScheduleEnabled,
	}}
	exec := &fakeExecutor{done: make(chan struct{})}
	current := start.Add(2 * time.Second)
	s, err := New(repo, exec, Config{
		PollInterval: time.Hour,
		Now:          func() time.Time { return current },
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := s.enqueueDue(context.Background()); err != nil {
		t.Fatal(err)
	}
	schedule := <-s.jobs
	s.run(context.Background(), schedule)

	repo.mu.Lock()
	next := repo.next
	repo.mu.Unlock()
	if next.IsZero() {
		t.Fatal("next_run_at tidak diperbarui")
	}
	if next != start.Add(2500*time.Millisecond) {
		t.Fatalf("next_run_at = %s, want %s", next, start.Add(2500*time.Millisecond))
	}
}
