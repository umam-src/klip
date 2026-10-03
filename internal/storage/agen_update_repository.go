package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/umam-src/klip/internal/domain"
)

// UpdateAgenWithParent memperbarui identitas, konfigurasi AI, status, dan parent Agen.
// Parent harus berada pada Ruang yang sama dan tidak boleh membentuk siklus.
func (r *Repository) UpdateAgenWithParent(ctx context.Context, agent domain.Agen) error {
	if err := validateAgen(agent); err != nil {
		return fmt.Errorf("agen: %w", err)
	}

	if _, err := r.GetAgenWithParent(ctx, agent.ID); err != nil {
		return err
	}
	if err := r.validateAgenParent(ctx, agent); err != nil {
		return err
	}

	role := strings.TrimSpace(agent.Role)
	if role == "" {
		role = "Agen"
	}
	status := agent.Status
	if status == "" {
		status = domain.AgenStatusActive
	}
	var parentID any
	if agent.ParentID != nil {
		parentID = string(*agent.ParentID)
	}

	result, err := r.db.ExecContext(ctx, `UPDATE agen SET parent_id = ?, name = ?, role = ?, description = ?, provider_id = ?, model_id = ?, status = ?, updated_at = ? WHERE id = ? AND ruang_id = ?`, parentID, agent.Name, role, agent.Description, agent.ProviderID, agent.ModelID, status, time.Now().UTC().Format(time.RFC3339Nano), agent.ID, agent.RuangID)
	if err != nil {
		return fmt.Errorf("ubah agen: %w", err)
	}
	if affected, err := result.RowsAffected(); err != nil {
		return fmt.Errorf("cek perubahan agen: %w", err)
	} else if affected != 1 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) validateAgenParent(ctx context.Context, agent domain.Agen) error {
	if agent.ParentID == nil {
		return nil
	}

	parentID := strings.TrimSpace(string(*agent.ParentID))
	if parentID == "" || parentID == string(agent.ID) {
		return fmt.Errorf("agen: parent tidak valid: %w", ErrInvalid)
	}

	var parentRuangID string
	if err := r.db.QueryRowContext(ctx, `SELECT ruang_id FROM agen WHERE id = ?`, parentID).Scan(&parentRuangID); errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("agen: parent tidak ditemukan: %w", ErrInvalid)
	} else if err != nil {
		return fmt.Errorf("cek parent agen: %w", err)
	} else if parentRuangID != string(agent.RuangID) {
		return fmt.Errorf("agen: parent tidak sesuai dengan ruang: %w", ErrInvalid)
	}

	visited := map[string]struct{}{string(agent.ID): {}}
	current := parentID
	for {
		if _, ok := visited[current]; ok {
			return fmt.Errorf("agen: hierarki parent membentuk siklus: %w", ErrInvalid)
		}
		visited[current] = struct{}{}

		var next sql.NullString
		if err := r.db.QueryRowContext(ctx, `SELECT parent_id FROM agen WHERE id = ?`, current).Scan(&next); errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("agen: parent tidak ditemukan: %w", ErrInvalid)
		} else if err != nil {
			return fmt.Errorf("baca hierarki parent agen: %w", err)
		}
		if !next.Valid || strings.TrimSpace(next.String) == "" {
			return nil
		}
		current = strings.TrimSpace(next.String)
	}
}
