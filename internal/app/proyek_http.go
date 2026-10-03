package app

import (
	"net/http"
	"strings"

	"github.com/umam-src/klip/internal/domain"
)

type createProyekRequest struct {
	ID          string `json:"id"`
	GoalID      string `json:"goal_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

type createNativeTugasRequest struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id,omitempty"`
	Title    string `json:"title"`
	Status   string `json:"status,omitempty"`
}

type createHasilKerjaRequest struct {
	ID          string `json:"id"`
	ExecutionID string `json:"execution_id,omitempty"`
	TugasID     string `json:"tugas_id,omitempty"`
	Kind        string `json:"kind"`
	Name        string `json:"name"`
	Path        string `json:"path"`
}

func (a *App) handleProyek(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan belum siap")
		return
	}
	ruangID := domain.ID(r.PathValue("id"))
	if strings.TrimSpace(string(ruangID)) == "" {
		writeError(w, http.StatusNotFound, "ruang tidak ditemukan")
		return
	}

	switch r.Method {
	case http.MethodGet:
		items, err := a.repo.ListProyekByRuang(r.Context(), ruangID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var req createProyekRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		status := domain.Status(strings.TrimSpace(req.Status))
		if status == "" {
			status = domain.StatusDraft
		}
		proyek := domain.Proyek{
			ID:          domain.ID(strings.TrimSpace(req.ID)),
			RuangID:     ruangID,
			GoalID:      domain.ID(strings.TrimSpace(req.GoalID)),
			Title:       strings.TrimSpace(req.Title),
			Description: strings.TrimSpace(req.Description),
			Status:      status,
		}
		if err := a.repo.CreateProyek(r.Context(), proyek); err != nil {
			writeStorageError(w, err)
			return
		}
		created, err := a.repo.GetProyek(r.Context(), proyek.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "proyek berhasil dibuat tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func (a *App) handleProyekChild(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan belum siap")
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/proyek/"), "/"), "/")
	if len(parts) != 2 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
		return
	}

	proyekID := domain.ID(parts[0])
	switch parts[1] {
	case "tugas":
		a.handleProyekTugas(w, r, proyekID)
	case "hasil":
		a.handleProyekHasil(w, r, proyekID)
	case "aktivitas":
		a.handleAktivitas(w, r, proyekID)
	case "komentar":
		a.handleProyekKomentar(w, r, proyekID)
	default:
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
	}
}

func (a *App) handleProyekTugas(w http.ResponseWriter, r *http.Request, proyekID domain.ID) {
	switch r.Method {
	case http.MethodGet:
		items, err := a.repo.ListTugasByProyek(r.Context(), proyekID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var req createNativeTugasRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		proyek, err := a.repo.GetProyek(r.Context(), proyekID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		status := domain.Status(strings.TrimSpace(req.Status))
		if status == "" {
			status = domain.StatusDraft
		}
		var parentID *domain.ID
		if value := strings.TrimSpace(req.ParentID); value != "" {
			id := domain.ID(value)
			parentID = &id
		}
		tugas := domain.Tugas{
			ID:       domain.ID(strings.TrimSpace(req.ID)),
			RuangID:  proyek.RuangID,
			ProyekID: proyek.ID,
			ParentID: parentID,
			Title:    strings.TrimSpace(req.Title),
			Status:   status,
		}
		if err := a.repo.CreateTugasNative(r.Context(), tugas); err != nil {
			writeStorageError(w, err)
			return
		}
		created, err := a.repo.GetTugasNative(r.Context(), tugas.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "tugas berhasil dibuat tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func (a *App) handleProyekHasil(w http.ResponseWriter, r *http.Request, proyekID domain.ID) {
	switch r.Method {
	case http.MethodGet:
		items, err := a.repo.ListHasilKerjaByProyek(r.Context(), proyekID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var req createHasilKerjaRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		var executionID, tugasID *domain.ID
		if value := strings.TrimSpace(req.ExecutionID); value != "" {
			id := domain.ID(value)
			executionID = &id
		}
		if value := strings.TrimSpace(req.TugasID); value != "" {
			id := domain.ID(value)
			tugasID = &id
		}
		hasil := domain.HasilKerja{
			ID:          domain.ID(strings.TrimSpace(req.ID)),
			ProyekID:    proyekID,
			TugasID:     tugasID,
			ExecutionID: executionID,
			Kind:        strings.TrimSpace(req.Kind),
			Name:        strings.TrimSpace(req.Name),
			Path:        strings.TrimSpace(req.Path),
		}
		proyek, err := a.repo.GetProyek(r.Context(), proyekID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		hasil.RuangID = proyek.RuangID
		if err := a.repo.CreateHasilKerja(r.Context(), hasil); err != nil {
			writeStorageError(w, err)
			return
		}
		created, err := a.repo.GetHasilKerja(r.Context(), hasil.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "hasil kerja berhasil dibuat tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}
