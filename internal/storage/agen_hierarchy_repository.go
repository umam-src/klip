package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/umam-src/klip/internal/domain"
)

func (r *Repository) CreateAgenWithParent(ctx context.Context, agen domain.Agen) error {
	if err := validateIDName(agen.ID, agen.Name); err != nil {
		return fmt.Errorf("agen: %w", err)
	}
	if strings.TrimSpace(string(agen.RuangID)) == "" {
		return fmt.Errorf("agen: ruang wajib diisi: %w", ErrInvalid)
	}
	var parentID any
	if agen.ParentID != nil {
		value := strings.TrimSpace(string(*agen.ParentID))
		if value == "" || value == string(agen.ID) {
			return fmt.Errorf("agen: parent tidak valid: %w", ErrInvalid)
		}
		var parentRuangID string
		err := r.db.QueryRowContext(ctx, `SELECT ruang_id FROM agen WHERE id = ?`, value).Scan(&parentRuangID)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("agen: parent tidak ditemukan: %w", ErrInvalid)
		}
		if err != nil {
			return fmt.Errorf("cek parent agen: %w", err)
		}
		if parentRuangID != string(agen.RuangID) {
			return fmt.Errorf("agen: parent tidak sesuai dengan ruang: %w", ErrInvalid)
		}
		parentID = value
	}
	created, updated := timestamps(agen.CreatedAt, agen.UpdatedAt)
	_, err := r.db.ExecContext(ctx, `INSERT INTO agen (id, ruang_id, parent_id, name, description, provider_id, model_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, agen.ID, agen.RuangID, parentID, agen.Name, agen.Description, agen.ProviderID, agen.ModelID, created, updated)
	if err != nil {
		return fmt.Errorf("buat agen: %w", err)
	}
	return nil
}

func (r *Repository) GetAgenWithParent(ctx context.Context, id domain.ID) (domain.Agen, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Agen{}, fmt.Errorf("agen: %w", ErrInvalid)
	}
	var agen domain.Agen
	var parentID sql.NullString
	var created, updated string
	err := r.db.QueryRowContext(ctx, `SELECT id, ruang_id, parent_id, name, description, provider_id, model_id, created_at, updated_at FROM agen WHERE id = ?`, id).Scan(&agen.ID, &agen.RuangID, &parentID, &agen.Name, &agen.Description, &agen.ProviderID, &agen.ModelID, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Agen{}, ErrNotFound
	}
	if err != nil {
		return domain.Agen{}, fmt.Errorf("ambil agen: %w", err)
	}
	if parentID.Valid {
		value := domain.ID(parentID.String)
		agent.ParentID = &value
	}
	var parseErr error
	if agen.CreatedAt, parseErr = parseTime(created); parseErr != nil {
		return domain.Agen{}, parseErr
	}
	if agen.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
		return domain.Agen{}, parseErr
	}
	return agen, nil
}

func (r *Repository) ListAgenByRuangWithParent(ctx context.Context, ruangID domain.ID) ([]domain.Agen, error) {
	if strings.TrimSpace(string(ruangID)) == "" {
		return nil, fmt.Errorf("agen: ruang wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, ruang_id, parent_id, name, description, provider_id, model_id, created_at, updated_at FROM agen WHERE ruang_id = ? ORDER BY parent_id IS NOT NULL, parent_id, created_at, id`, ruangID)
	if err != nil {
		return nil, fmt.Errorf("daftar agen: %w", err)
	}
	defer rows.Close()

	var result []domain.Agen
	for rows.Next() {
		var agen domain.Agen
		var parentID sql.NullString
		var created, updated string
		if err := rows.Scan(&agen.ID, &agen.RuangID, &parentID, &agen.Name, &agen.Description, &agen.ProviderID, &agen.ModelID, &created, &updated); err != nil {
			return nil, fmt.Errorf("baca agen: %w", err)
		}
		if parentID.Valid {
			value := domain.ID(parentID.String)
			agent.ParentID = &value
		}
		var parseErr error
		if agen.CreatedAt, parseErr = parseTime(created); parseErr != nil {
			return nil, parseErr
		}
		if agen.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
			return nil, parseErr
		}
		result = append(result, agen)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar agen: %w", err)
	}
	return result, nil
}
