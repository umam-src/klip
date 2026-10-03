package app

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

const defaultActivityLimit = 50

type activityResponse struct {
	ID        domain.ID  `json:"id"`
	TugasID   *domain.ID `json:"tugas_id,omitempty"`
	AgenID    *domain.ID `json:"agen_id,omitempty"`
	Type      string     `json:"type"`
	Summary   string     `json:"summary,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

func (a *App) handleAktivitas(w http.ResponseWriter, r *http.Request, proyekID domain.ID) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
		return
	}

	limit, err := parseActivityLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "limit harus berupa angka 1 sampai 100")
		return
	}

	events, err := a.repo.ListEventsByProyek(r.Context(), proyekID, limit)
	if err != nil {
		writeStorageError(w, err)
		return
	}

	items := make([]activityResponse, 0, len(events))
	for _, event := range events {
		items = append(items, activityResponse{
			ID:        event.ID,
			TugasID:   event.TugasID,
			AgenID:    event.AgenID,
			Type:      event.Type,
			Summary:   activitySummary(event.Type),
			CreatedAt: event.CreatedAt,
		})
	}
	writeJSON(w, http.StatusOK, items)
}

func parseActivityLimit(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultActivityLimit, nil
	}
	limit, err := strconv.Atoi(value)
	if err != nil || limit < 1 || limit > 100 {
		return 0, errInvalidActivityLimit{}
	}
	return limit, nil
}

type errInvalidActivityLimit struct{}

func (errInvalidActivityLimit) Error() string { return "limit di luar rentang" }

func activitySummary(eventType string) string {
	switch eventType {
	case domain.EventExecutionStarted:
		return "Eksekusi dimulai"
	case domain.EventExecutionCompleted:
		return "Eksekusi selesai"
	case domain.EventExecutionFailed:
		return "Eksekusi gagal"
	case domain.EventExecutionCancelled:
		return "Eksekusi dibatalkan"
	default:
		return ""
	}
}
