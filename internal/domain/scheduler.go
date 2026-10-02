package domain

import "time"

type ScheduleStatus string

const (
	ScheduleEnabled  ScheduleStatus = "enabled"
	ScheduleDisabled ScheduleStatus = "disabled"
)

type Schedule struct {
	ID          ID
	Name        string
	PekerjaanID ID
	TugasID     *ID
	AgenID      ID
	Program     string
	Arguments   []string
	Interval    time.Duration
	NextRunAt   time.Time
	Status      ScheduleStatus
	RetryLimit  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ScheduleRunStatus string

const (
	ScheduleRunQueued    ScheduleRunStatus = "queued"
	ScheduleRunRunning   ScheduleRunStatus = "running"
	ScheduleRunSucceeded ScheduleRunStatus = "succeeded"
	ScheduleRunFailed    ScheduleRunStatus = "failed"
	ScheduleRunSkipped   ScheduleRunStatus = "skipped"
)

type ScheduleRun struct {
	ID         ID
	ScheduleID ID
	Status     ScheduleRunStatus
	Attempt    int
	StartedAt  time.Time
	FinishedAt *time.Time
	Error      string
}
