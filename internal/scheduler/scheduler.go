package scheduler

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/umam-src/klip/internal/agent"
	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

var ErrClosed = errors.New("scheduler sudah dihentikan")

type Executor interface {
	Execute(context.Context, agent.ExecutionRequest) (agent.ExecutionResult, error)
}

type Config struct {
	PollInterval time.Duration
	Workers      int
	Heartbeat    time.Duration
	RetryDelay   time.Duration
	Now          func() time.Time
}

type Scheduler struct {
	repo     storage.SchedulerRepository
	executor Executor
	cfg      Config
	jobs     chan domain.Schedule
	stop     chan struct{}
	done     chan struct{}
	mu       sync.Mutex
	active   map[domain.ID]struct{}
	online   map[domain.ID]time.Time
}

func New(repo storage.SchedulerRepository, executor Executor, cfg Config) (*Scheduler, error) {
	if repo == nil || executor == nil {
		return nil, fmt.Errorf("scheduler: repository dan executor wajib diisi")
	}
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = time.Second
	}
	if cfg.Workers <= 0 {
		cfg.Workers = 1
	}
	if cfg.Heartbeat <= 0 {
		cfg.Heartbeat = 30 * time.Second
	}
	if cfg.RetryDelay <= 0 {
		cfg.RetryDelay = time.Second
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return &Scheduler{
		repo:     repo,
		executor: executor,
		cfg:      cfg,
		jobs:     make(chan domain.Schedule, cfg.Workers*2),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		active:   make(map[domain.ID]struct{}),
		online:   make(map[domain.ID]time.Time),
	}, nil
}

func (s *Scheduler) Start(ctx context.Context) {
	go s.loop(ctx)
}

func (s *Scheduler) Stop() {
	select {
	case <-s.stop:
		return
	default:
		close(s.stop)
	}
	<-s.done
}

func (s *Scheduler) loop(ctx context.Context) {
	var wg sync.WaitGroup
	for i := 0; i < s.cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.worker(ctx)
		}()
	}
	ticker := time.NewTicker(s.cfg.PollInterval)
	defer ticker.Stop()
	defer func() {
		wg.Wait()
		close(s.done)
	}()

	_ = s.enqueueDue(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stop:
			return
		case now := <-ticker.C:
			s.heartbeat(now)
			_ = s.enqueueDue(ctx)
		}
	}
}

func (s *Scheduler) worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stop:
			return
		case schedule := <-s.jobs:
			s.run(ctx, schedule)
		}
	}
}

func (s *Scheduler) enqueueDue(ctx context.Context) error {
	schedules, err := s.repo.ListSchedules(ctx)
	if err != nil {
		return err
	}
	now := s.cfg.Now().UTC()
	for _, schedule := range schedules {
		if schedule.Status != domain.ScheduleEnabled || schedule.NextRunAt.After(now) {
			continue
		}
		s.mu.Lock()
		_, active := s.active[schedule.ID]
		if !active {
			s.active[schedule.ID] = struct{}{}
		}
		s.mu.Unlock()
		if active {
			continue
		}
		select {
		case s.jobs <- schedule:
		default:
			s.mu.Lock()
			delete(s.active, schedule.ID)
			s.mu.Unlock()
		}
	}
	return nil
}

func (s *Scheduler) run(ctx context.Context, schedule domain.Schedule) {
	defer func() {
		s.mu.Lock()
		delete(s.active, schedule.ID)
		s.mu.Unlock()
	}()

	started := s.cfg.Now().UTC()
	attempts := schedule.RetryLimit + 1
	for attempt := 1; attempt <= attempts; attempt++ {
		run := domain.ScheduleRun{
			ID:         fmt.Sprintf("%s-%d-%d", schedule.ID, started.UnixNano(), attempt),
			ScheduleID: schedule.ID,
			Status:     domain.ScheduleRunRunning,
			Attempt:    attempt,
			StartedAt:  started,
		}
		_ = s.repo.CreateScheduleRun(ctx, run)
		_, err := s.executor.Execute(ctx, agent.ExecutionRequest{
			PekerjaanID: schedule.PekerjaanID,
			TugasID:     schedule.TugasID,
			AgenID:      schedule.AgenID,
			Program:     schedule.Program,
			Arguments:   schedule.Arguments,
		})
		finished := s.cfg.Now().UTC()
		if err == nil {
			_ = s.repo.FinishScheduleRun(ctx, run.ID, domain.ScheduleRunSucceeded, "", finished)
			break
		}
		_ = s.repo.FinishScheduleRun(ctx, run.ID, domain.ScheduleRunFailed, err.Error(), finished)
		if attempt < attempts {
			select {
			case <-ctx.Done():
				return
			case <-s.stop:
				return
			case <-time.After(s.cfg.RetryDelay):
			}
		}
	}

	next := schedule.NextRunAt.Add(schedule.Interval)
	now := s.cfg.Now().UTC()
	for !next.After(now) {
		next = next.Add(schedule.Interval)
	}
	_ = s.repo.UpdateScheduleRun(ctx, schedule.ID, next)
}

func (s *Scheduler) heartbeat(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id := range s.active {
		s.online[id] = now.UTC()
	}
}

func (s *Scheduler) Heartbeats() map[domain.ID]time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	result := make(map[domain.ID]time.Time, len(s.online))
	for id, at := range s.online {
		result[id] = at
	}
	return result
}
