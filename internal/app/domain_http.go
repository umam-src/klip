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

type createPekerjaanRequest struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status,omitempty"`
	SasaranID string `json:"sasaran_id,omitempty"`
}

type createTugasRequest struct {
	ID       string `json:"id"`
	ParentID string `json:"parent_id,omitempty"`
	Title    string `json:"title"`
	Status   string `json:"status,omitempty"`
	Position int    `json:"position"`
}

type createHasilRequest struct {
	ID      string `json:"id"`
	TugasID string `json:"tugas_id,omitempty"`
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Path    string `json:"path"`
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
	case "pekerjaan":
		a.handlePekerjaan(w, r, ruangID)
	default:
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
	}
}

func (a *App) handlePekerjaanChild(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan belum siap")
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/pekerjaan/"), "/"), "/")
	if len(parts) != 2 || parts[0] == "" {
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
		return
	}

	pekerjaanID := domain.ID(parts[0])
	switch parts[1] {
	case "tugas":
		a.handleTugas(w, r, pekerjaanID)
	case "hasil":
		a.handleHasil(w, r, pekerjaanID)
	case "aktivitas":
		a.handleAktivitas(w, r, pekerjaanID)
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
		pekerjaan := domain.Pekerjaan{
			ID:        domain.ID(strings.TrimSpace(req.ID)),
			RuangID:   ruangID,
			SasaranID: sasaranID,
			Title:     strings.TrimSpace(req.Title),
			Status:    status,
		}
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

func (a *App) handleTugas(w http.ResponseWriter, r *http.Request, pekerjaanID domain.ID) {
	switch r.Method {
	case http.MethodGet:
		items, err := a.repo.ListTugasByPekerjaan(r.Context(), pekerjaanID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var req createTugasRequest
		if !decodeJSON(w, r, &req) {
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
			ID:          domain.ID(strings.TrimSpace(req.ID)),
			PekerjaanID: pekerjaanID,
			ParentID:    parentID,
			Title:       strings.TrimSpace(req.Title),
			Status:      status,
			Position:    req.Position,
		}
		if err := a.repo.CreateTugas(r.Context(), tugas); err != nil {
			writeStorageError(w, err)
			return
		}
		created, err := a.repo.GetTugas(r.Context(), tugas.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "tugas berhasil dibuat tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func (a *App) handleHasil(w http.ResponseWriter, r *http.Request, pekerjaanID domain.ID) {
	switch r.Method {
	case http.MethodGet:
		items, err := a.repo.ListHasilByPekerjaan(r.Context(), pekerjaanID)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var req createHasilRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		var tugasID *domain.ID
		if value := strings.TrimSpace(req.TugasID); value != "" {
			id := domain.ID(value)
			tugasID = &id
		}
		hasil := domain.Hasil{
			ID:          domain.ID(strings.TrimSpace(req.ID)),
			PekerjaanID: pekerjaanID,
			TugasID:     tugasID,
			Kind:        strings.TrimSpace(req.Kind),
			Name:        strings.TrimSpace(req.Name),
			Path:        strings.TrimSpace(req.Path),
		}
		if err := validateHasil(hasil); err != nil {
			writeStorageError(w, err)
			return
		}
		if hasil.TugasID != nil {
			tugas, err := a.repo.GetTugas(r.Context(), *hasil.TugasID)
			if err != nil {
				writeStorageError(w, err)
				return
			}
			if tugas.PekerjaanID != pekerjaanID {
				writeError(w, http.StatusBadRequest, "tugas tidak sesuai dengan pekerjaan")
				return
			}
		}
		if err := a.repo.CreateHasil(r.Context(), hasil); err != nil {
			writeStorageError(w, err)
			return
		}
		created, err := a.repo.GetHasil(r.Context(), hasil.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "hasil berhasil dibuat tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
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
