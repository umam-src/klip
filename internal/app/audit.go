package app

import (
	"context"
	"log/slog"

	"github.com/umam-src/klip/internal/domain"
)

// auditAction mencatat aksi penting tanpa memasukkan alasan, isi prompt, atau data rahasia.
// Kegagalan audit tidak membatalkan aksi yang sudah berhasil dilakukan.
func (a *App) auditAction(ctx context.Context, pekerjaanID domain.ID, tugasID *domain.ID, eventType string) {
	if a.repo == nil || pekerjaanID == "" || eventType == "" {
		return
	}
	if err := a.repo.AppendEvent(ctx, domain.Event{
		PekerjaanID: pekerjaanID,
		TugasID:     tugasID,
		Type:        eventType,
	}); err != nil {
		slog.DebugContext(ctx, "gagal mencatat audit aksi", "event_type", eventType)
	}
}
