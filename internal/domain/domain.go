package domain

import "time"

type ID string

type Status string

const (
    StatusDraft     Status = "draft"
    StatusReady     Status = "ready"
    StatusRunning   Status = "running"
    StatusWaiting   Status = "waiting"
    StatusBlocked   Status = "blocked"
    StatusCompleted Status = "completed"
    StatusFailed    Status = "failed"
    StatusCancelled Status = "cancelled"
)

type Ruang struct {
    ID        ID
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Agen struct {
    ID          ID
    RuangID     ID
    Name        string
    Description string
    ProviderID  string
    ModelID     string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Sasaran struct {
    ID        ID
    RuangID   ID
    Title     string
    Status    Status
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Pekerjaan struct {
    ID        ID
    RuangID   ID
    SasaranID *ID
    Title     string
    Status    Status
    CreatedAt time.Time
    UpdatedAt time.Time
}

type Tugas struct {
    ID          ID
    PekerjaanID ID
    ParentID    *ID
    Title       string
    Status      Status
    Position    int
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

type Sesi struct {
    ID         ID
    PekerjaanID ID
    AgenID     ID
    Status     Status
    StartedAt  time.Time
    FinishedAt *time.Time
}

type Hasil struct {
    ID         ID
    PekerjaanID ID
    TugasID    *ID
    Kind       string
    Name       string
    Path       string
    CreatedAt  time.Time
}
