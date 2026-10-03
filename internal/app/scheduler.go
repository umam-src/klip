package app

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

type scheduleRequest struct {
	Name        string     `json:"name"`
	ProyekID    domain.ID  `json:"proyek_id"`
	TugasID     *domain.ID `json:"tugas_id,omitempty"`
	AgenID      domain.ID  `json:"agen_id"`
	Program     string     `json:"program"`
	Arguments   []string   `json:"arguments,omitempty"`
	IntervalSec int64      `json:"interval_seconds"`
	NextRunAt   string     `json:"next_run_at,omitempty"`
	RetryLimit  int        `json:"retry_limit,omitempty"`
}

type scheduleResponse struct {
	ID        domain.ID             `json:"id"`
	Name      string                `json:"name"`
	ProyekID  domain.ID             `json:"proyek_id"`
	TugasID   *domain.ID            `json:"tugas_id,omitempty"`
	AgenID    domain.ID             `json:"agen_id"`
	Program   string                `json:"program"`
	Arguments []string              `json:"arguments"`
	IntervalSec int64               `json:"interval_seconds"`
	NextRunAt time.Time             `json:"next_run_at"`
	Status    domain.ScheduleStatus `json:"status"`
	RetryLimit int                  `json:"retry_limit"`
}

func (a *App) handleScheduler(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "database belum siap")
		return
	}
	switch r.Method {
	case http.MethodGet:
		schedules, err := a.repo.ListSchedules(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "gagal membaca jadwal")
			return
		}
		out := make([]scheduleResponse, 0, len(schedules))
		for _, schedule := range schedules {
			out = append(out, toScheduleResponse(schedule))
		}
		writeJSON(w, http.StatusOK, out)
	case http.MethodPost:
		var req scheduleRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		schedule, err := a.newSchedule(req)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if err := a.repo.CreateSchedule(r.Context(), schedule); err != nil {
			writeError(w, http.StatusBadRequest, "jadwal tidak valid")
			return
		}
		a.auditAction(r.Context(), schedule.ProyekID, schedule.TugasID, domain.EventScheduleCreated)
		writeJSON(w, http.StatusCreated, toScheduleResponse(schedule))
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func (a *App) handleSchedulerChild(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 5 || parts[0] != "api" || parts[1] != "v1" || parts[2] != "scheduler" {
		writeError(w, http.StatusNotFound, "jadwal tidak ditemukan")
		return
	}
	id := domain.ID(parts[3])
	if parts[4] != "riwayat" || r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
		return
	}
	limit := 50
	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			writeError(w, http.StatusBadRequest, "limit tidak valid")
			return
		}
		limit = parsed
	}
	runs, err := a.repo.ListScheduleRuns(r.Context(), id, limit)
	if err != nil {
		writeError(w, http.StatusBadRequest, "riwayat jadwal tidak valid")
		return
	}
	writeJSON(w, http.StatusOK, runs)
}

func (a *App) newSchedule(req scheduleRequest) (domain.Schedule, error) {
	if strings.TrimSpace(req.Name) == "" || req.ProyekID == "" || req.AgenID == "" || strings.TrimSpace(req.Program) == "" {
		return domain.Schedule{}, errInvalidSchedule
	}
	if req.IntervalSec < 1 || req.IntervalSec > 7*24*60*60 {
		return domain.Schedule{}, errInvalidSchedule
	}
	if req.RetryLimit < 0 || req.RetryLimit > 10 {
		return domain.Schedule{}, errInvalidSchedule
	}
	next := time.Now().UTC().Add(time.Duration(req.IntervalSec) * time.Second)
	if strings.TrimSpace(req.NextRunAt) != "" {
		parsed, err := time.Parse(time.RFC3339, req.NextRunAt)
		if err != nil {
			return domain.Schedule{}, errInvalidSchedule
		}
		next = parsed.UTC()
	}
	now := time.Now().UTC()
	return domain.Schedule{
		ID:        newScheduleID(),
		Name:      strings.TrimSpace(req.Name),
		ProyekID:  req.ProyekID,
		TugasID:   req.TugasID,
		AgenID:    req.AgenID,
		Program:   strings.TrimSpace(req.Program),
		Arguments: append([]string(nil), req.Arguments...),
		Interval:  time.Duration(req.IntervalSec) * time.Second,
		NextRunAt: next,
		Status:    domain.ScheduleEnabled,
		RetryLimit: req.RetryLimit,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

var errInvalidSchedule = &scheduleError{"jadwal tidak valid"}

type scheduleError struct {
	message string
}

func (e *scheduleError) Error() string { return e.message }

func toScheduleResponse(s domain.Schedule) scheduleResponse {
	return scheduleResponse{
		ID:         s.ID,
		Name:       s.Name,
		ProyekID:   s.ProyekID,
		TugasID:    s.TugasID,
		AgenID:     s.AgenID,
		Program:    s.Program,
		Arguments:  s.Arguments,
		IntervalSec: int64(s.Interval / time.Second),
		NextRunAt:  s.NextRunAt,
		Status:     s.Status,
		RetryLimit: s.RetryLimit,
	}
}

func newScheduleID() domain.ID {
	var data [16]byte
	if _, err := rand.Read(data[:]); err == nil {
		return domain.ID(hex.EncodeToString(data[:]))
	}
	return domain.ID(strconv.FormatInt(time.Now().UnixNano(), 10))
}
