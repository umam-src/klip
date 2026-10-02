package app

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/umam-src/klip/internal/domain"
	"github.com/umam-src/klip/internal/storage"
)

const maxUploadedResultSize int64 = 16 << 20

func (a *App) handleHasilFile(w http.ResponseWriter, r *http.Request) {
	if a.repo == nil || a.resultStore == nil {
		writeError(w, http.StatusServiceUnavailable, "penyimpanan hasil belum siap")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "id hasil wajib diisi")
		return
	}
	hasil, err := a.repo.GetHasil(r.Context(), domain.ID(id))
	if errors.Is(err, storage.ErrNotFound) {
		writeError(w, http.StatusNotFound, "hasil tidak ditemukan")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal membaca hasil")
		return
	}

	switch r.Method {
	case http.MethodPut:
		if r.ContentLength > maxUploadedResultSize {
			writeError(w, http.StatusRequestEntityTooLarge, "berkas hasil terlalu besar")
			return
		}
		if err := a.resultStore.Put(hasil.Path, io.LimitReader(r.Body, maxUploadedResultSize+1)); err != nil {
			writeStorageError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	case http.MethodGet:
		data, err := a.resultStore.Read(hasil.Path)
		if errors.Is(err, storage.ErrNotFound) {
			writeError(w, http.StatusNotFound, "berkas hasil belum tersedia")
			return
		}
		if err != nil {
			writeStorageError(w, err)
			return
		}
		contentType := http.DetectContentType(data)
		if contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
		w.Header().Set("Content-Disposition", "inline; filename=\""+strings.ReplaceAll(hasil.Name, "\"", "")+"\"")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	default:
		w.Header().Set("Allow", "GET, PUT")
		writeError(w, http.StatusMethodNotAllowed, "metode tidak didukung")
	}
}
