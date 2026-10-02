package app

import (
	"net/http"
	"strings"

	"github.com/umam-src/klip/internal/domain"
)

type createSasaranRequest struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status,omitempty"`
}

func (a *App) handleSasaran(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan belum siap")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/ruang/")
	id = strings.TrimSuffix(id, "/sasaran")
	if id == "" || strings.Contains(id, "/") {
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
		return
	}
	ruangID := domain.ID(id)

	switch r.Method {
	case http.MethodGet:
		items, err := a.repo.ListSasaranByRuang(r.Context(), ruangID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var req createSasaranRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		status := domain.Status(strings.TrimSpace(req.Status))
		if status == "" {
			status = domain.StatusDraft
		}
		sasaran := domain.Sasaran{
			ID:      domain.ID(strings.TrimSpace(req.ID)),
			RuangID: ruangID,
			Title:   strings.TrimSpace(req.Title),
			Status:  status,
		}
		if err := a.repo.CreateSasaran(r.Context(), sasaran); err != nil {
			writeStorageError(w, err)
			return
		}
		created, err := a.repo.GetSasaran(r.Context(), sasaran.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "sasaran berhasil dibuat tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}
