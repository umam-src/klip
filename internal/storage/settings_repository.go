package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type AISettings struct {
	Provider string
	BaseURL  string
	Model    string
}

func (r *Repository) GetAISettings(ctx context.Context) (AISettings, bool, error) {
	var settings AISettings
	err := r.db.QueryRowContext(ctx, `SELECT provider, base_url, model FROM pengaturan_ai WHERE id = 1`).Scan(&settings.Provider, &settings.BaseURL, &settings.Model)
	if errors.Is(err, sql.ErrNoRows) {
		return AISettings{}, false, nil
	}
	if err != nil {
		return AISettings{}, false, fmt.Errorf("ambil pengaturan AI: %w", err)
	}
	return settings, true, nil
}

func (r *Repository) SaveAISettings(ctx context.Context, settings AISettings) error {
	settings.Provider = strings.TrimSpace(settings.Provider)
	settings.BaseURL = strings.TrimSpace(settings.BaseURL)
	settings.Model = strings.TrimSpace(settings.Model)
	if settings.Provider == "" || settings.BaseURL == "" || settings.Model == "" {
		return fmt.Errorf("pengaturan AI: provider, alamat, dan model wajib diisi: %w", ErrInvalid)
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO pengaturan_ai (id, provider, base_url, model, updated_at)
		VALUES (1, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			provider = excluded.provider,
			base_url = excluded.base_url,
			model = excluded.model,
			updated_at = excluded.updated_at`,
		settings.Provider, settings.BaseURL, settings.Model, time.Now().UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("simpan pengaturan AI: %w", err)
	}
	return nil
}
