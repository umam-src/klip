package app

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

type assignTugasAgenRequest struct {
	AgenID string `json:"agen_id"`
}

type tugasAgenResponse struct {
	TugasID    domain.ID `json:"tugas_id"`
	AgenID     domain.ID `json:"agen_id"`
	AssignedAt string    `json:"assigned_at"`
}

func (a *App) handleTugasAgen(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan belum siap")
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/tugas/"), "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "agen" {
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
		return
	}

	tugasID := domain.ID(parts[0])
	if _, err := a.repo.GetTugasNative(r.Context(), tugasID); err != nil {
		writeStorageError(w, err)
		return
	}

	switch r.Method {
	case http.MethodGet:
		assignment, err := a.repo.GetTugasAssignment(r.Context(), tugasID)
		if err != nil {
			if err == storage.ErrNotFound {
				writeJSON(w, http.StatusOK, nil)
				return
			}
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, tugasAgenResponse{
			TugasID:    assignment.TugasID,
			AgenID:     assignment.AgenID,
			AssignedAt: assignment.AssignedAt.UTC().Format(time.RFC3339Nano),
		})
	case http.MethodPut:
		var req assignTugasAgenRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "format JSON tidak valid")
			return
		}
		agentID := domain.ID(strings.TrimSpace(req.AgenID))
		if agentID == "" {
			writeError(w, http.StatusBadRequest, "agen_id wajib diisi")
			return
		}
		if err := a.repo.AssignTugasToAgen(r.Context(), tugasID, agentID, time.Now().UTC()); err != nil {
			writeStorageError(w, err)
			return
		}
		assignment, err := a.repo.GetTugasAssignment(r.Context(), tugasID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "assignment berhasil disimpan tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusOK, tugasAgenResponse{
			TugasID:    assignment.TugasID,
			AgenID:     assignment.AgenID,
			AssignedAt: assignment.AssignedAt.UTC().Format(time.RFC3339Nano),
		})
	case http.MethodDelete:
		if err := a.repo.UnassignTugas(r.Context(), tugasID); err != nil {
			writeStorageError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}
