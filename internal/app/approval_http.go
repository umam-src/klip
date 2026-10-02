package app

import (
	"net/http"
	"strings"

	"github.com/umam-src/klip/internal/domain"
)

type createApprovalRequest struct {
	ID     string `json:"id"`
	Reason string `json:"reason,omitempty"`
}

type decideApprovalRequest struct {
	Reason string `json:"reason,omitempty"`
}

func (a *App) handlePekerjaanApproval(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan belum siap")
		return
	}
	id := domain.ID(strings.TrimPrefix(r.URL.Path, "/api/v1/pekerjaan/"))
	id = domain.ID(strings.TrimSuffix(string(id), "/approval"))
	if strings.TrimSpace(string(id)) == "" {
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := a.repo.ListApprovalsByPekerjaan(r.Context(), id)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var req createApprovalRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		approval := domain.Approval{ID: domain.ID(strings.TrimSpace(req.ID)), PekerjaanID: id, Reason: strings.TrimSpace(req.Reason), Status: domain.ApprovalPending}
		if err := a.repo.CreateApproval(r.Context(), approval); err != nil {
			writeStorageError(w, err)
			return
		}
		created, err := a.repo.GetApproval(r.Context(), approval.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "approval berhasil dibuat tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func (a *App) handleTugasApproval(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan belum siap")
		return
	}
	id := domain.ID(strings.TrimPrefix(r.URL.Path, "/api/v1/tugas/"))
	id = domain.ID(strings.TrimSuffix(string(id), "/approval"))
	if strings.TrimSpace(string(id)) == "" {
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
		return
	}
	if _, err := a.repo.GetTugas(r.Context(), id); err != nil {
		writeStorageError(w, err)
		return
	}
	switch r.Method {
	case http.MethodGet:
		items, err := a.repo.ListApprovalsByTugas(r.Context(), id)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, items)
	case http.MethodPost:
		var req createApprovalRequest
		if !decodeJSON(w, r, &req) {
			return
		}
		tugas, err := a.repo.GetTugas(r.Context(), id)
		if err != nil {
			writeStorageError(w, err)
			return
		}
		approval := domain.Approval{ID: domain.ID(strings.TrimSpace(req.ID)), PekerjaanID: tugas.PekerjaanID, TugasID: &id, Reason: strings.TrimSpace(req.Reason), Status: domain.ApprovalPending}
		if err := a.repo.CreateApproval(r.Context(), approval); err != nil {
			writeStorageError(w, err)
			return
		}
		created, err := a.repo.GetApproval(r.Context(), approval.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "approval berhasil dibuat tetapi gagal dibaca")
			return
		}
		writeJSON(w, http.StatusCreated, created)
	default:
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}

func (a *App) handleApprovalDecision(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan belum siap")
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v1/approval/"), "/"), "/")
	if len(parts) != 2 || parts[0] == "" || (parts[1] != "approve" && parts[1] != "reject") {
		writeError(w, http.StatusNotFound, "jalur tidak ditemukan")
		return
	}
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
		return
	}
	var req decideApprovalRequest
	if r.ContentLength != 0 && !decodeJSON(w, r, &req) {
		return
	}
	status := domain.ApprovalApproved
	if parts[1] == "reject" {
		status = domain.ApprovalRejected
	}
	if err := a.repo.DecideApproval(r.Context(), domain.ID(parts[0]), status, req.Reason); err != nil {
		writeStorageError(w, err)
		return
	}
	approval, err := a.repo.GetApproval(r.Context(), domain.ID(parts[0]))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "approval berhasil diputuskan tetapi gagal dibaca")
		return
	}
	writeJSON(w, http.StatusOK, approval)
}
