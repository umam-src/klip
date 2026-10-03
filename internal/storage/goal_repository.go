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

// GoalRepository menyediakan batas penyimpanan Goal native.
type GoalRepository interface {
	CreateGoal(ctx context.Context, goal domain.Goal) error
	GetGoal(ctx context.Context, id domain.ID) (domain.Goal, error)
	ListGoalByRuang(ctx context.Context, ruangID domain.ID) ([]domain.Goal, error)
	UpdateGoal(ctx context.Context, goal domain.Goal) error
}

func (r *Repository) CreateGoal(ctx context.Context, goal domain.Goal) error {
	if err := validateGoal(goal); err != nil {
		return fmt.Errorf("goal: %w", err)
	}
	if err := r.validateGoalParent(ctx, goal); err != nil {
		return err
	}
	created, updated := timestamps(goal.CreatedAt, goal.UpdatedAt)
	_, err := r.db.ExecContext(ctx, `INSERT INTO goal (id, ruang_id, parent_goal_id, title, description, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, goal.ID, goal.RuangID, nullableID(goal.ParentGoalID), goal.Title, goal.Description, goal.Status, created, updated)
	if err != nil {
		return fmt.Errorf("buat goal: %w", err)
	}
	return nil
}

func (r *Repository) GetGoal(ctx context.Context, id domain.ID) (domain.Goal, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Goal{}, fmt.Errorf("goal: %w", ErrInvalid)
	}
	return r.scanGoal(r.db.QueryRowContext(ctx, `SELECT id, ruang_id, parent_goal_id, title, description, status, created_at, updated_at FROM goal WHERE id = ?`, id))
}

func (r *Repository) ListGoalByRuang(ctx context.Context, ruangID domain.ID) ([]domain.Goal, error) {
	if strings.TrimSpace(string(ruangID)) == "" {
		return nil, fmt.Errorf("goal: ruang wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, ruang_id, parent_goal_id, title, description, status, created_at, updated_at FROM goal WHERE ruang_id = ? ORDER BY created_at, id`, ruangID)
	if err != nil {
		return nil, fmt.Errorf("daftar goal: %w", err)
	}
	defer rows.Close()

	var result []domain.Goal
	for rows.Next() {
		goal, err := r.scanGoal(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, goal)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar goal: %w", err)
	}
	return result, nil
}

func (r *Repository) UpdateGoal(ctx context.Context, goal domain.Goal) error {
	if err := validateGoal(goal); err != nil {
		return fmt.Errorf("goal: %w", err)
	}
	current, err := r.GetGoal(ctx, goal.ID)
	if err != nil {
		return err
	}
	if current.RuangID != goal.RuangID {
		return fmt.Errorf("goal: ruang tidak sesuai: %w", ErrInvalid)
	}
	if !current.Status.CanTransitionTo(goal.Status) {
		return fmt.Errorf("goal: transisi status tidak valid: %w", ErrInvalid)
	}
	if err := r.validateGoalParent(ctx, goal); err != nil {
		return err
	}
	updated := goal.UpdatedAt
	if updated.IsZero() {
		updated = time.Now().UTC()
	}
	_, err = r.db.ExecContext(ctx, `UPDATE goal SET parent_goal_id = ?, title = ?, description = ?, status = ?, updated_at = ? WHERE id = ?`, nullableID(goal.ParentGoalID), goal.Title, goal.Description, goal.Status, updated.UTC().Format(time.RFC3339Nano), goal.ID)
	if err != nil {
		return fmt.Errorf("ubah goal: %w", err)
	}
	return nil
}

func (r *Repository) validateGoalParent(ctx context.Context, goal domain.Goal) error {
	if goal.ParentGoalID == nil {
		return nil
	}
	if *goal.ParentGoalID == goal.ID {
		return fmt.Errorf("goal: induk tidak boleh dirinya sendiri: %w", ErrInvalid)
	}
	parent, err := r.GetGoal(ctx, *goal.ParentGoalID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return fmt.Errorf("goal: induk tidak ditemukan: %w", ErrInvalid)
		}
		return err
	}
	if parent.RuangID != goal.RuangID {
		return fmt.Errorf("goal: induk tidak sesuai dengan ruang: %w", ErrInvalid)
	}
	return nil
}

func validateGoal(goal domain.Goal) error {
	if strings.TrimSpace(string(goal.ID)) == "" || strings.TrimSpace(string(goal.RuangID)) == "" || strings.TrimSpace(goal.Title) == "" || !goal.Status.IsKnown() {
		return ErrInvalid
	}
	return nil
}

type goalScanner interface {
	Scan(dest ...any) error
}

func (r *Repository) scanGoal(row goalScanner) (domain.Goal, error) {
	var goal domain.Goal
	var parentID sql.NullString
	var created, updated string
	if err := row.Scan(&goal.ID, &goal.RuangID, &parentID, &goal.Title, &goal.Description, &goal.Status, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Goal{}, ErrNotFound
		}
		return domain.Goal{}, fmt.Errorf("ambil goal: %w", err)
	}
	if parentID.Valid {
		id := domain.ID(parentID.String)
		goal.ParentGoalID = &id
	}
	var err error
	if goal.CreatedAt, err = parseTime(created); err != nil {
		return domain.Goal{}, err
	}
	if goal.UpdatedAt, err = parseTime(updated); err != nil {
		return domain.Goal{}, err
	}
	return goal, nil
}

func nullableID(id *domain.ID) any {
	if id == nil {
		return nil
	}
	return *id
}
