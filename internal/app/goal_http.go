package app

import (
	"net/http"
	"strings"

	"github.com/umam-src/klip/internal/domain"
)

type createGoalRequest struct {
	ID           string `json:"id"`
	ParentGoalID string `json:"parent_goal_id,omitempty"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	Status       string `json:"status,omitempty"`
}

type updateGoalRequest struct {
	ParentGoalID string `json:"parent_goal_id,omitempty"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	Status       string `json:"status"`
}

func (a *App) handleGoal(w http.ResponseWriter, r *http.Request, ruangID domain.ID) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan belum siap")
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := a.repo.ListGoalByRuang(r.Context(), ruangID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var req createGoalRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		goal := domain.Goal{
			ID:           domain.ID(strings.TrimSpace(req.ID)),
			RuangID:      ruangID,
			ParentGoalID: parseOptionalID(req.ParentGoalID),
			Title:        strings.TrimSpace(req.Title),
			Description:  strings.TrimSpace(req.Description),
			Status:       parseGoalStatus(req.Status),
		}
		if err := a.repo.CreateGoal(r.Context(), goal); err != nil {
			writeStorageError(w, err)
			return
		}
		created, err := a.repo.GetGoal(r.Context(), goal.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "goal berhasil dibuat tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func (a *App) handleGoalDetail(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan belum siap")
		return
	}
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/goal/"), "/")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
		return
	}
	goalID := domain.ID(id)

	switch r.Method {
	case http.MethodGet:
		goal, err := a.repo.GetGoal(r.Context(), goalID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, goal)
	case http.MethodPut:
		current, err := a.repo.GetGoal(r.Context(), goalID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		var req updateGoalRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		status := domain.GoalStatus(strings.TrimSpace(req.Status))
		if status == "" {
			status = current.Status
		}
		updated := domain.Goal{
			ID:           current.ID,
			RuangID:      current.RuangID,
			ParentGoalID: parseOptionalID(req.ParentGoalID),
			Title:        strings.TrimSpace(req.Title),
			Description:  strings.TrimSpace(req.Description),
			Status:       status,
			CreatedAt:    current.CreatedAt,
			UpdatedAt:    current.UpdatedAt,
		}
		if err := a.repo.UpdateGoal(r.Context(), updated); err != nil {
			writeStorageError(w, err)
			return
		}
		result, err := a.repo.GetGoal(r.Context(), goalID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "goal berhasil diubah tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusOK, result)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func parseOptionalID(value string) *domain.ID {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	id := domain.ID(value)
	return &id
}

func parseGoalStatus(value string) domain.GoalStatus {
	status := domain.GoalStatus(strings.TrimSpace(value))
	if status == "" {
		return domain.GoalStatusActive
	}
	return status
}
