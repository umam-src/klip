package scheduler

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/umam-src/klip/internal/agent"
	"github.com/umam-src/klip/internal/domain"
)

type fakeRepo struct {
	mu       sync.Mutex
	schedule domain.Schedule
	runs     []domain.ScheduleRun
	next     time.Time
}

func (r *fakeRepo) CreateSchedule(context.Context, domain.Schedule) error { return nil }

func (r *fakeRepo) GetSchedule(context.Context, domain.ID) (domain.Schedule, error) {
	return r.schedule, nil
}

func (r *fakeRepo) ListSchedules(context.Context) ([]domain.Schedule, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	schedule := r.schedule
	if !r.next.IsZero() {
		schedule.NextRunAt = r.next
	}
	return []domain.Schedule{schedule}, nil
}

func (r *fakeRepo) UpdateScheduleRun(_ context.Context, _ domain.ID, next time.Time) error {
	r.mu.Lock()
	r.next = next
	r.mu.Unlock()
	return nil
}

func (r *fakeRepo) CreateScheduleRun(_ context.Context, run domain.ScheduleRun) error {
	r.mu.Lock()
	r.runs = append(r.runs, run)
	r.mu.Unlock()
	return nil
}

func (r *fakeRepo) FinishScheduleRun(_ context.Context, id domain.ID, status domain.ScheduleRunStatus, errText string, finished time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.runs {
		if r.runs[i].ID == id {
			r.runs[i].Status = status
			runsFinished := finished
			r.runs[i].FinishedAt = &runsFinished
			r.runs[i].Error = errText
		}
	}
	return nil
}

func (r *fakeRepo) ListScheduleRuns(context.Context, domain.ID, int) ([]domain.ScheduleRun, error) {
	return nil, nil
}

type fakeExecutor struct {
	mu      sync.Mutex
	calls   int
	fail    int
	done    chan struct{}
	request agent.ExecutionRequest
}

func (e *fakeExecutor) Execute(_ context.Context, request agent.ExecutionRequest) (agent.ExecutionResult, error) {
	e.mu.Lock()
	e.calls++
	calls := e.calls
	e.request = request
	e.mu.Unlock()
	if calls <= e.fail {
		return agent.ExecutionResult{}, errors.New("gagal")
	}
	if e.done != nil {
		close(e.done)
		e.done = nil
	}
	return agent.ExecutionResult{}, nil
}

func TestSchedulerRetryAndHistory(t *testing.T) {
	repo := &fakeRepo{schedule: domain.Schedule{
		ID:         "s1",
		Name:       "uji",
		RuangID:    "r1",
		ProyekID:   "p1",
		AgenID:     "a1",
		Program:    "true",
		Interval:   time.Hour,
		NextRunAt:  time.Now().Add(-time.Second),
		Status:     domain.ScheduleEnabled,
		RetryLimit: 2,
	}}
	exec := &fakeExecutor{fail: 1, done: make(chan struct{})}
	s, err := New(repo, exec, Config{PollInterval: 5 * time.Millisecond, RetryDelay: time.Millisecond, Heartbeat: 5 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)
	select {
	case <-exec.done:
	case <-time.After(time.Second):
		t.Fatal("scheduler tidak menjalankan jadwal")
	}
	s.Stop()
	repo.mu.Lock()
	defer repo.mu.Unlock()
	if exec.calls != 2 {
		t.Fatalf("calls = %d, want 2", exec.calls)
	}
	if len(repo.runs) != 2 {
		t.Fatalf("history = %d, want 2", len(repo.runs))
	}
	for _, run := range repo.runs {
		if run.Status != domain.ScheduleRunSucceeded && run.Status != domain.ScheduleRunFailed {
			t.Fatalf("status = %q", run.Status)
		}
	}
}

func TestSchedulerPassesTaskAssignmentToExecutor(t *testing.T) {
	tugasID := domain.ID("t1")
	repo := &fakeRepo{schedule: domain.Schedule{
		ID:         "s-assignment",
		Name:       "uji assignment",
		RuangID:    "r1",
		ProyekID:   "p1",
		TugasID:    &tugasID,
		AgenID:     "a1",
		Program:    "true",
		Interval:   time.Hour,
		NextRunAt:  time.Now().Add(-time.Second),
		Status:     domain.ScheduleEnabled,
	}}
	exec := &fakeExecutor{done: make(chan struct{})}
	s, err := New(repo, exec, Config{PollInterval: 5 * time.Millisecond})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.Start(ctx)
	select {
	case <-exec.done:
	case <-time.After(time.Second):
		t.Fatal("scheduler tidak menjalankan tugas terjadwal")
	}
	s.Stop()

	exec.mu.Lock()
	request := exec.request
	exec.mu.Unlock()
	if request.ProyekID != "p1" {
		t.Fatalf("proyek_id = %q, want p1", request.ProyekID)
	}
	if request.TugasID == nil || *request.TugasID != tugasID {
		t.Fatalf("tugas_id = %v, want %q", request.TugasID, tugasID)
	}
	if request.AgenID != "a1" {
		t.Fatalf("agen_id = %q, want a1", request.AgenID)
	}
}
