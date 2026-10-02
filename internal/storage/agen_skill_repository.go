package storage

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/umam-src/klip/internal/domain"
)

type AgenSkillRepository interface {
	AssignSkill(ctx context.Context, relation domain.AgenSkill) error
	RemoveSkill(ctx context.Context, relation domain.AgenSkill) error
	ListSkillsByAgen(ctx context.Context, agenID domain.ID) ([]domain.AgenSkill, error)
}

func (r *Repository) AssignSkill(ctx context.Context, relation domain.AgenSkill) error {
	if err := validateAgenSkill(relation); err != nil {
		return err
	}
	if _, err := r.GetAgen(ctx, relation.AgenID); err != nil {
		return fmt.Errorf("skill agen: %w", err)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO agen_skill (agen_id, skill_name) VALUES (?, ?)`, relation.AgenID, relation.SkillName)
	if err != nil {
		return fmt.Errorf("pasang skill: %w", err)
	}
	return nil
}

func (r *Repository) RemoveSkill(ctx context.Context, relation domain.AgenSkill) error {
	if err := validateAgenSkill(relation); err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, `DELETE FROM agen_skill WHERE agen_id = ? AND skill_name = ?`, relation.AgenID, relation.SkillName)
	if err != nil {
		return fmt.Errorf("lepas skill: %w", err)
	}
	if affected, err := result.RowsAffected(); err == nil && affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ListSkillsByAgen(ctx context.Context, agenID domain.ID) ([]domain.AgenSkill, error) {
	if strings.TrimSpace(string(agenID)) == "" {
		return nil, fmt.Errorf("skill agen: agen wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT agen_id, skill_name FROM agen_skill WHERE agen_id = ? ORDER BY skill_name`, agenID)
	if err != nil {
		return nil, fmt.Errorf("daftar skill agen: %w", err)
	}
	defer rows.Close()

	result := make([]domain.AgenSkill, 0)
	for rows.Next() {
		var relation domain.AgenSkill
		if err := rows.Scan(&relation.AgenID, &relation.SkillName); err != nil {
			return nil, fmt.Errorf("baca skill agen: %w", err)
		}
		result = append(result, relation)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar skill agen: %w", err)
	}
	return result, nil
}

func validateAgenSkill(relation domain.AgenSkill) error {
	if strings.TrimSpace(string(relation.AgenID)) == "" {
		return fmt.Errorf("skill agen: agen wajib diisi: %w", ErrInvalid)
	}
	if !domain.IsValidSkillName(relation.SkillName) {
		return fmt.Errorf("skill agen: nama skill tidak valid: %w", ErrInvalid)
	}
	return nil
}

var _ sql.Result
var _ AgenSkillRepository = (*Repository)(nil)
