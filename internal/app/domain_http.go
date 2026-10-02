package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

type createRuangRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type createAgenRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ProviderID  string `json:"provider_id,omitempty"`
	ModelID     string `json:"model_id,omitempty"`
}

type createPekerjaanRequest struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status,omitempty"`
	SasaranID string `json:"sasaran_id,omitempty"`
}

func (a *App) handleRuang(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan belum siap")
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := a.repo.ListRuang(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "gagal membaca ruang")
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var req createRuangRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		ruang := domain.Ruang{ID: domain.ID(strings.TrimSpace(req.ID)), Name: strings.TrimSpace(req.Name)}
		if err := a.repo.CreateRuang(r.Context(), ruang); err != nil {
			writeStorageError(w, err)
			return
		}
		created, err := a.repo.GetRuang(r.Context(), ruang.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "ruang berhasil dibuat tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func (a *App) handleRuangChild(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan belum siap")
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/ruang/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) != 2 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
		return
	}

	ruangID := domain.ID(parts[0])
	switch parts[1] {
	case "agen":
		a.handleAgen(w, r, ruangID)
	case "pekerjaan":
		a.handlePekerjaan(w, r, ruangID)
	default:
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
	}
}

func (a *App) handleAgen(w http.ResponseWriter, r *http.Request, ruangID domain.ID) {
	switch r.Method {
	case http.MethodGet:
		items, err := a.repo.ListAgenByRuang(r.Context(), ruangID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var req createAgenRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		agen := domain.Agen{
			ID: domain.ID(strings.TrimSpace(req.ID)), RuangID: ruangID,
			Name: strings.TrimSpace(req.Name), Description: strings.TrimSpace(req.Description),
			ProviderID: strings.TrimSpace(req.ProviderID), ModelID: strings.TrimSpace(req.ModelID),
		}
		if err := a.repo.CreateAgen(r.Context(), agen); err != nil {
			writeStorageError(w, err)
			return
		}
		created, err := a.repo.GetAgen(r.Context(), agen.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "agen berhasil dibuat tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func (a *App) handlePekerjaan(w http.ResponseWriter, r *http.Request, ruangID domain.ID) {
	switch r.Method {
	case http.MethodGet:
		items, err := a.repo.ListPekerjaanByRuang(r.Context(), ruangID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var req createPekerjaanRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		status := domain.Status(strings.TrimSpace(req.Status))
		if status == "" {
			status = domain.StatusDraft
		}
		var sasaranID *domain.ID
		if value := strings.TrimSpace(req.SasaranID); value != "" {
			id := domain.ID(value)
			sasaranID = &id
		}
		pekerjaan := domain.Pekerjaan{ID: domain.ID(strings.TrimSpace(req.ID)), RuangID: ruangID, SasaranID: sasaranID, Title: strings.TrimSpace(req.Title), Status: status}
		if err := a.repo.CreatePekerjaan(r.Context(), pekerjaan); err != nil {
			writeStorageError(w, err)
			return
		}
		created, err := a.repo.GetPekerjaan(r.Context(), pekerjaan.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "pekerjaan berhasil dibuat tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, value any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	if err := decoder.Decode(value); err != nil {
		writeError(w, http.StatusBadRequest, "permintaan tidak valid")
		return false
	}
	return true
}

func writeStorageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, storage.ErrInvalid):
		writeError(w, http.StatusBadRequest, "data tidak valid")
	case errors.Is(err, storage.ErrNotFound):
		writeError(w, http.StatusNotFound, "data tidak ditemukan")
	default:
		writeError(w, http.StatusInternalServerError, "gagal menyimpan data")
	}
}

var _ = time.Time{}
