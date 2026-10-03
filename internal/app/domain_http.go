package app

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

type createRuangRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type createAgenRequest struct {
	ID          string `json:"id"`
	ParentID    string `json:"parent_id,omitempty"`
	Name        string `json:"name"`
	Role        string `json:"role,omitempty"`
	Description string `json:"description,omitempty"`
	ProviderID  string `json:"provider_id,omitempty"`
	ModelID     string `json:"model_id,omitempty"`
	Status      string `json:"status,omitempty"`
}

type updateAgenRequest struct {
	ParentID    string `json:"parent_id,omitempty"`
	Name        string `json:"name"`
	Role        string `json:"role"`
	Description string `json:"description,omitempty"`
	ProviderID  string `json:"provider_id,omitempty"`
	ModelID     string `json:"model_id,omitempty"`
	Status      string `json:"status,omitempty"`
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
		ruang := domain.Ruang{
			ID:   domain.ID(strings.TrimSpace(req.ID)),
			Name: strings.TrimSpace(req.Name),
		}
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
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/ruang/"), "/"), "/")
	if len(parts) != 2 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
		return
	}

	ruangID := domain.ID(parts[0])
	switch parts[1] {
	case "agen":
		a.handleAgen(w, r, ruangID)
	default:
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
	}
}

func (a *App) handleAgen(w http.ResponseWriter, r *http.Request, ruangID domain.ID) {
	agentID := strings.TrimSpace(r.URL.Query().Get("agent_id"))
	switch r.Method {
	case http.MethodGet:
		if agentID != "" {
			agent, err := a.repo.GetAgenWithParent(r.Context(), domain.ID(agentID))
			if err != nil {
				writeStorageError(w, err)
				return
			}
			if agent.RuangID != ruangID {
				writeStorageError(w, storage.ErrNotFound)
				return
			}
			writeJSON(w, http.StatusOK, agent)
			return
		}
		items, err := a.repo.ListAgenByRuangWithParent(r.Context(), ruangID)
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
		var parentID *domain.ID
		if value := strings.TrimSpace(req.ParentID); value != "" {
			id := domain.ID(value)
			parentID = &id
		}
		status := domain.AgenStatus(strings.TrimSpace(req.Status))
		if status == "" {
			status = domain.AgenStatusActive
		}
		agent := domain.Agen{
			ID:          domain.ID(strings.TrimSpace(req.ID)),
			RuangID:     ruangID,
			ParentID:    parentID,
			Name:        strings.TrimSpace(req.Name),
			Role:        strings.TrimSpace(req.Role),
			Description: strings.TrimSpace(req.Description),
			ProviderID:  strings.TrimSpace(req.ProviderID),
			ModelID:     strings.TrimSpace(req.ModelID),
			Status:      status,
		}
		if err := a.repo.CreateAgenWithParent(r.Context(), agent); err != nil {
			writeStorageError(w, err)
			return
		}
		created, err := a.repo.GetAgenWithParent(r.Context(), agent.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "agen berhasil dibuat tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	case http.MethodPut:
		if agentID == "" {
			writeError(w, http.StatusBadRequest, "agent_id wajib diisi")
			return
		}
		a.updateAgen(w, r, domain.ID(agentID), ruangID)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func (a *App) updateAgen(w http.ResponseWriter, r *http.Request, id, ruangID domain.ID) {
	current, err := a.repo.GetAgenWithParent(r.Context(), id)
	if err != nil {
		writeStorageError(w, err)
		return
	}
	if current.RuangID != ruangID {
		writeStorageError(w, storage.ErrNotFound)
		return
	}

	var req updateAgenRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	var parentID *domain.ID
	if value := strings.TrimSpace(req.ParentID); value != "" {
		parent := domain.ID(value)
		parentID = &parent
	}
	status := domain.AgenStatus(strings.TrimSpace(req.Status))
	if status == "" {
		status = current.Status
	}
	updated := domain.Agen{
		ID:          id,
		RuangID:     ruangID,
		ParentID:    parentID,
		Name:        strings.TrimSpace(req.Name),
		Role:        strings.TrimSpace(req.Role),
		Description: strings.TrimSpace(req.Description),
		ProviderID:  strings.TrimSpace(req.ProviderID),
		ModelID:     strings.TrimSpace(req.ModelID),
		Status:      status,
		CreatedAt:   current.CreatedAt,
		UpdatedAt:   current.UpdatedAt,
	}
	if err := a.repo.UpdateAgenWithParent(r.Context(), updated); err != nil {
		writeStorageError(w, err)
		return
	}
	result, err := a.repo.GetAgenWithParent(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "agen berhasil diubah tetapi gagal dibaca")
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func decodeJSON(w http.ResponseWriter, r *http.Request, value any) bool {
	if contentType := strings.TrimSpace(r.Header.Get("Content-Type")); contentType != "" {
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil || !strings.EqualFold(mediaType, "application/json") {
			writeError(w, http.StatusUnsupportedMediaType, "content-type harus application/json")
			return false
		}
	}

	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	if err := decoder.Decode(value); err != nil {
		writeError(w, http.StatusBadRequest, "permintaan tidak valid")
		return false
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		writeError(w, http.StatusBadRequest, "permintaan memiliki data tambahan")
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
