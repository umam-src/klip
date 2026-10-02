package app

import (
	"net/http"
	"strings"

	"github.com/umam-src/klip/internal/domain"
)

type createKomentarRequest struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id,omitempty"`
	Body     string `json:"body"`
}

func (a *App) handleKomentar(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan belum siap")
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/pekerjaan/"), "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "komentar" {
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
		return
	}
	pekerjaanID := domain.ID(parts[0])
	switch r.Method {
	case http.MethodGet:
		items, err := a.repo.ListKomentarByPekerjaan(r.Context(), pekerjaanID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		a.createKomentar(w, r, pekerjaanID, nil)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func (a *App) handleTugasKomentar(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan belum siap")
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/tugas/"), "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "komentar" {
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
		return
	}
	tugasID := domain.ID(parts[0])
	tugas, err := a.repo.GetTugas(r.Context(), tugasID)
	if err != nil {
		writeStorageError(w, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := a.repo.ListKomentarByTugas(r.Context(), tugasID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		a.createKomentar(w, r, tugas.PekerjaanID, &tugasID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func (a *App) createKomentar(w http.ResponseWriter, r *http.Request, pekerjaanID domain.ID, tugasID *domain.ID) {
	var req createKomentarRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	var parentID *domain.ID
	if value := strings.TrimSpace(req.ParentID); value != "" {
		id := domain.ID(value)
		parentID = &id
	}
	komentar := domain.Komentar{
		ID:          domain.ID(strings.TrimSpace(req.ID)),
		PekerjaanID: pekerjaanID,
		TugasID:     tugasID,
		ParentID:    parentID,
		Body:        strings.TrimSpace(req.Body),
	}
	if err := a.repo.CreateKomentar(r.Context(), komentar); err != nil {
		writeStorageError(w, err)
		return
	}
	created, err := a.repo.GetKomentar(r.Context(), komentar.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "komentar berhasil dibuat tetapi gagal dibaca")
		return
	}
	writeJSON(w, http.StatusCreated, created)
}
