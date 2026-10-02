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

var (
	ErrNotFound = errors.New("data tidak ditemukan")
	ErrInvalid  = errors.New("data tidak valid")
)

type RuangRepository interface {
	CreateRuang(ctx context.Context, ruang domain.Ruang) error
	GetRuang(ctx context.Context, id domain.ID) (domain.Ruang, error)
	ListRuang(ctx context.Context) ([]domain.Ruang, error)
}

type AgenRepository interface {
	CreateAgen(ctx context.Context, agen domain.Agen) error
	GetAgen(ctx context.Context, id domain.ID) (domain.Agen, error)
	ListAgenByRuang(ctx context.Context, ruangID domain.ID) ([]domain.Agen, error)
}

type PekerjaanRepository interface {
	CreatePekerjaan(ctx context.Context, pekerjaan domain.Pekerjaan) error
	GetPekerjaan(ctx context.Context, id domain.ID) (domain.Pekerjaan, error)
	ListPekerjaanByRuang(ctx context.Context, ruangID domain.ID) ([]domain.Pekerjaan, error)
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateRuang(ctx context.Context, ruang domain.Ruang) error {
	if err := validateIDName(ruang.ID, ruang.Name); err != nil {
		return fmt.Errorf("ruang: %w", err)
	}
	created, updated := timestamps(ruang.CreatedAt, ruang.UpdatedAt)
	_, err := r.db.ExecContext(ctx, `INSERT INTO ruang (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`, ruang.ID, ruang.Name, created, updated)
	if err != nil {
		return fmt.Errorf("buat ruang: %w", err)
	}
	return nil
}

func (r *Repository) GetRuang(ctx context.Context, id domain.ID) (domain.Ruang, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Ruang{}, fmt.Errorf("ruang: %w", ErrInvalid)
	}
	var ruang domain.Ruang
	var created, updated string
	err := r.db.QueryRowContext(ctx, `SELECT id, name, created_at, updated_at FROM ruang WHERE id = ?`, id).Scan(&ruang.ID, &ruang.Name, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Ruang{}, ErrNotFound
	}
	if err != nil {
		return domain.Ruang{}, fmt.Errorf("ambil ruang: %w", err)
	}
	var parseErr error
	if ruang.CreatedAt, parseErr = parseTime(created); parseErr != nil {
		return domain.Ruang{}, parseErr
	}
	if ruang.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
		return domain.Ruang{}, parseErr
	}
	return ruang, nil
}

func (r *Repository) ListRuang(ctx context.Context) ([]domain.Ruang, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, created_at, updated_at FROM ruang ORDER BY created_at, id`)
	if err != nil {
		return nil, fmt.Errorf("daftar ruang: %w", err)
	}
	defer rows.Close()

	var result []domain.Ruang
	for rows.Next() {
		var ruang domain.Ruang
		var created, updated string
		if err := rows.Scan(&ruang.ID, &ruang.Name, &created, &updated); err != nil {
			return nil, fmt.Errorf("baca ruang: %w", err)
		}
		var parseErr error
		if ruang.CreatedAt, parseErr = parseTime(created); parseErr != nil {
			return nil, parseErr
		}
		if ruang.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
			return nil, parseErr
		}
		result = append(result, ruang)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar ruang: %w", err)
	}
	return result, nil
}

func (r *Repository) CreateAgen(ctx context.Context, agen domain.Agen) error {
	if err := validateIDName(agen.ID, agen.Name); err != nil {
		return fmt.Errorf("agen: %w", err)
	}
	if strings.TrimSpace(string(agen.RuangID)) == "" {
		return fmt.Errorf("agen: ruang wajib diisi: %w", ErrInvalid)
	}
	created, updated := timestamps(agen.CreatedAt, agen.UpdatedAt)
	_, err := r.db.ExecContext(ctx, `INSERT INTO agen (id, ruang_id, name, description, provider_id, model_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, agen.ID, agen.RuangID, agen.Name, agen.Description, agen.ProviderID, agen.ModelID, created, updated)
	if err != nil {
		return fmt.Errorf("buat agen: %w", err)
	}
	return nil
}

func (r *Repository) GetAgen(ctx context.Context, id domain.ID) (domain.Agen, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Agen{}, fmt.Errorf("agen: %w", ErrInvalid)
	}
	var agen domain.Agen
	var created, updated string
	err := r.db.QueryRowContext(ctx, `SELECT id, ruang_id, name, description, provider_id, model_id, created_at, updated_at FROM agen WHERE id = ?`, id).Scan(&agen.ID, &agen.RuangID, &agen.Name, &agen.Description, &agen.ProviderID, &agen.ModelID, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Agen{}, ErrNotFound
	}
	if err != nil {
		return domain.Agen{}, fmt.Errorf("ambil agen: %w", err)
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

func (r *Repository) ListAgenByRuang(ctx context.Context, ruangID domain.ID) ([]domain.Agen, error) {
	if strings.TrimSpace(string(ruangID)) == "" {
		return nil, fmt.Errorf("agen: ruang wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, ruang_id, name, description, provider_id, model_id, created_at, updated_at FROM agen WHERE ruang_id = ? ORDER BY created_at, id`, ruangID)
	if err != nil {
		return nil, fmt.Errorf("daftar agen: %w", err)
	}
	defer rows.Close()

	var result []domain.Agen
	for rows.Next() {
		var agen domain.Agen
		var created, updated string
		if err := rows.Scan(&agen.ID, &agen.RuangID, &agen.Name, &agen.Description, &agen.ProviderID, &agen.ModelID, &created, &updated); err != nil {
			return nil, fmt.Errorf("baca agen: %w", err)
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

func (r *Repository) CreatePekerjaan(ctx context.Context, pekerjaan domain.Pekerjaan) error {
	if err := validateIDName(pekerjaan.ID, pekerjaan.Title); err != nil {
		return fmt.Errorf("pekerjaan: %w", err)
	}
	if strings.TrimSpace(string(pekerjaan.RuangID)) == "" {
		return fmt.Errorf("pekerjaan: ruang wajib diisi: %w", ErrInvalid)
	}
	created, updated := timestamps(pekerjaan.CreatedAt, pekerjaan.UpdatedAt)
	var sasaranID any
	if pekerjaan.SasaranID != nil {
		sasaranID = string(*pekerjaan.SasaranID)
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO pekerjaan (id, ruang_id, sasaran_id, title, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, pekerjaan.ID, pekerjaan.RuangID, sasaranID, pekerjaan.Title, pekerjaan.Status, created, updated)
	if err != nil {
		return fmt.Errorf("buat pekerjaan: %w", err)
	}
	return nil
}

func (r *Repository) GetPekerjaan(ctx context.Context, id domain.ID) (domain.Pekerjaan, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Pekerjaan{}, fmt.Errorf("pekerjaan: %w", ErrInvalid)
	}
	var pekerjaan domain.Pekerjaan
	var sasaranID sql.NullString
	var created, updated string
	err := r.db.QueryRowContext(ctx, `SELECT id, ruang_id, sasaran_id, title, status, created_at, updated_at FROM pekerjaan WHERE id = ?`, id).Scan(&pekerjaan.ID, &pekerjaan.RuangID, &sasaranID, &pekerjaan.Title, &pekerjaan.Status, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Pekerjaan{}, ErrNotFound
	}
	if err != nil {
		return domain.Pekerjaan{}, fmt.Errorf("ambil pekerjaan: %w", err)
	}
	if sasaranID.Valid {
		value := domain.ID(sasaranID.String)
		pekerjaan.SasaranID = &value
	}
	var parseErr error
	if pekerjaan.CreatedAt, parseErr = parseTime(created); parseErr != nil {
		return domain.Pekerjaan{}, parseErr
	}
	if pekerjaan.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
		return domain.Pekerjaan{}, parseErr
	}
	return pekerjaan, nil
}

func (r *Repository) ListPekerjaanByRuang(ctx context.Context, ruangID domain.ID) ([]domain.Pekerjaan, error) {
	if strings.TrimSpace(string(ruangID)) == "" {
		return nil, fmt.Errorf("pekerjaan: ruang wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, ruang_id, sasaran_id, title, status, created_at, updated_at FROM pekerjaan WHERE ruang_id = ? ORDER BY created_at, id`, ruangID)
	if err != nil {
		return nil, fmt.Errorf("daftar pekerjaan: %w", err)
	}
	defer rows.Close()

	var result []domain.Pekerjaan
	for rows.Next() {
		var pekerjaan domain.Pekerjaan
		var sasaranID sql.NullString
		var created, updated string
		if err := rows.Scan(&pekerjaan.ID, &pekerjaan.RuangID, &sasaranID, &pekerjaan.Title, &pekerjaan.Status, &created, &updated); err != nil {
			return nil, fmt.Errorf("baca pekerjaan: %w", err)
		}
		if sasaranID.Valid {
			value := domain.ID(sasaranID.String)
			pekerjaan.SasaranID = &value
		}
		var parseErr error
		if pekerjaan.CreatedAt, parseErr = parseTime(created); parseErr != nil {
			return nil, parseErr
		}
		if pekerjaan.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
			return nil, parseErr
		}
		result = append(result, pekerjaan)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar pekerjaan: %w", err)
	}
	return result, nil
}

func validateIDName(id domain.ID, name string) error {
	if strings.TrimSpace(string(id)) == "" || strings.TrimSpace(name) == "" {
		return ErrInvalid
	}
	return nil
}

func timestamps(created, updated time.Time) (string, string) {
	if created.IsZero() {
		created = time.Now().UTC()
	}
	if updated.IsZero() {
		updated = created
	}
	return created.UTC().Format(time.RFC3339Nano), updated.UTC().Format(time.RFC3339Nano)
}

func parseTime(value string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("waktu tidak valid: %w", err)
	}
	return parsed, nil
}
