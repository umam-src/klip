package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path"
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

type TugasRepository interface {
	CreateTugas(ctx context.Context, tugas domain.Tugas) error
	GetTugas(ctx context.Context, id domain.ID) (domain.Tugas, error)
	ListTugasByPekerjaan(ctx context.Context, pekerjaanID domain.ID) ([]domain.Tugas, error)
}

type HasilRepository interface {
	CreateHasil(ctx context.Context, hasil domain.Hasil) error
	GetHasil(ctx context.Context, id domain.ID) (domain.Hasil, error)
	ListHasilByPekerjaan(ctx context.Context, pekerjaanID domain.ID) ([]domain.Hasil, error)
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

func (r *Repository) CreateTugas(ctx context.Context, tugas domain.Tugas) error {
	if err := validateIDName(tugas.ID, tugas.Title); err != nil {
		return fmt.Errorf("tugas: %w", err)
	}
	if strings.TrimSpace(string(tugas.PekerjaanID)) == "" || tugas.Position < 0 {
		return fmt.Errorf("tugas: pekerjaan wajib diisi dan posisi tidak boleh negatif: %w", ErrInvalid)
	}
	if tugas.ParentID != nil {
		parentID := strings.TrimSpace(string(*tugas.ParentID))
		if parentID == "" {
			return fmt.Errorf("tugas: parent tidak valid: %w", ErrInvalid)
		}
		var parentPekerjaanID string
		err := r.db.QueryRowContext(ctx, `SELECT pekerjaan_id FROM tugas WHERE id = ?`, parentID).Scan(&parentPekerjaanID)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("tugas: parent tidak ditemukan: %w", ErrInvalid)
		}
		if err != nil {
			return fmt.Errorf("cek parent tugas: %w", err)
		}
		if parentPekerjaanID != string(tugas.PekerjaanID) {
			return fmt.Errorf("tugas: parent tidak sesuai dengan pekerjaan: %w", ErrInvalid)
		}
	}
	var parentID any
	if tugas.ParentID != nil {
		parentID = string(*tugas.ParentID)
	}
	created, updated := timestamps(tugas.CreatedAt, tugas.UpdatedAt)
	_, err := r.db.ExecContext(ctx, `INSERT INTO tugas (id, pekerjaan_id, parent_id, title, status, position, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, tugas.ID, tugas.PekerjaanID, parentID, tugas.Title, tugas.Status, tugas.Position, created, updated)
	if err != nil {
		return fmt.Errorf("buat tugas: %w", err)
	}
	return nil
}

func (r *Repository) GetTugas(ctx context.Context, id domain.ID) (domain.Tugas, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Tugas{}, fmt.Errorf("tugas: %w", ErrInvalid)
	}
	var tugas domain.Tugas
	var parentID sql.NullString
	var created, updated string
	err := r.db.QueryRowContext(ctx, `SELECT id, pekerjaan_id, parent_id, title, status, position, created_at, updated_at FROM tugas WHERE id = ?`, id).Scan(&tugas.ID, &tugas.PekerjaanID, &parentID, &tugas.Title, &tugas.Status, &tugas.Position, &created, &updated)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Tugas{}, ErrNotFound
	}
	if err != nil {
		return domain.Tugas{}, fmt.Errorf("ambil tugas: %w", err)
	}
	if parentID.Valid {
		value := domain.ID(parentID.String)
		tugas.ParentID = &value
	}
	var parseErr error
	if tugas.CreatedAt, parseErr = parseTime(created); parseErr != nil {
		return domain.Tugas{}, parseErr
	}
	if tugas.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
		return domain.Tugas{}, parseErr
	}
	return tugas, nil
}

func (r *Repository) ListTugasByPekerjaan(ctx context.Context, pekerjaanID domain.ID) ([]domain.Tugas, error) {
	if strings.TrimSpace(string(pekerjaanID)) == "" {
		return nil, fmt.Errorf("tugas: pekerjaan wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, pekerjaan_id, parent_id, title, status, position, created_at, updated_at FROM tugas WHERE pekerjaan_id = ? ORDER BY position, created_at, id`, pekerjaanID)
	if err != nil {
		return nil, fmt.Errorf("daftar tugas: %w", err)
	}
	defer rows.Close()

	var result []domain.Tugas
	for rows.Next() {
		var tugas domain.Tugas
		var parentID sql.NullString
		var created, updated string
		if err := rows.Scan(&tugas.ID, &tugas.PekerjaanID, &parentID, &tugas.Title, &tugas.Status, &tugas.Position, &created, &updated); err != nil {
			return nil, fmt.Errorf("baca tugas: %w", err)
		}
		if parentID.Valid {
			value := domain.ID(parentID.String)
			tugas.ParentID = &value
		}
		var parseErr error
		if tugas.CreatedAt, parseErr = parseTime(created); parseErr != nil {
			return nil, parseErr
		}
		if tugas.UpdatedAt, parseErr = parseTime(updated); parseErr != nil {
			return nil, parseErr
		}
		result = append(result, tugas)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar tugas: %w", err)
	}
	return result, nil
}

func (r *Repository) CreateHasil(ctx context.Context, hasil domain.Hasil) error {
	if err := validateHasil(hasil); err != nil {
		return fmt.Errorf("hasil: %w", err)
	}
	if hasil.TugasID != nil {
		tugasID := strings.TrimSpace(string(*hasil.TugasID))
		if tugasID == "" {
			return fmt.Errorf("hasil: tugas tidak valid: %w", ErrInvalid)
		}
		var tugasPekerjaanID string
		err := r.db.QueryRowContext(ctx, `SELECT pekerjaan_id FROM tugas WHERE id = ?`, tugasID).Scan(&tugasPekerjaanID)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("hasil: tugas tidak ditemukan: %w", ErrInvalid)
		}
		if err != nil {
			return fmt.Errorf("cek tugas hasil: %w", err)
		}
		if tugasPekerjaanID != string(hasil.PekerjaanID) {
			return fmt.Errorf("hasil: tugas tidak sesuai dengan pekerjaan: %w", ErrInvalid)
		}
	}
	var tugasID any
	if hasil.TugasID != nil {
		tugasID = string(*hasil.TugasID)
	}
	created := hasil.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO hasil (id, pekerjaan_id, tugas_id, kind, name, path, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)`, hasil.ID, hasil.PekerjaanID, tugasID, hasil.Kind, hasil.Name, hasil.Path, created.UTC().Format(time.RFC3339Nano))
	if err != nil {
		return fmt.Errorf("buat hasil: %w", err)
	}
	return nil
}

func (r *Repository) GetHasil(ctx context.Context, id domain.ID) (domain.Hasil, error) {
	if strings.TrimSpace(string(id)) == "" {
		return domain.Hasil{}, fmt.Errorf("hasil: %w", ErrInvalid)
	}
	var hasil domain.Hasil
	var tugasID sql.NullString
	var created string
	err := r.db.QueryRowContext(ctx, `SELECT id, pekerjaan_id, tugas_id, kind, name, path, created_at FROM hasil WHERE id = ?`, id).Scan(&hasil.ID, &hasil.PekerjaanID, &tugasID, &hasil.Kind, &hasil.Name, &hasil.Path, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Hasil{}, ErrNotFound
	}
	if err != nil {
		return domain.Hasil{}, fmt.Errorf("ambil hasil: %w", err)
	}
	if tugasID.Valid {
		value := domain.ID(tugasID.String)
		hasil.TugasID = &value
	}
	var parseErr error
	if hasil.CreatedAt, parseErr = parseTime(created); parseErr != nil {
		return domain.Hasil{}, parseErr
	}
	return hasil, nil
}

func (r *Repository) ListHasilByPekerjaan(ctx context.Context, pekerjaanID domain.ID) ([]domain.Hasil, error) {
	if strings.TrimSpace(string(pekerjaanID)) == "" {
		return nil, fmt.Errorf("hasil: pekerjaan wajib diisi: %w", ErrInvalid)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id, pekerjaan_id, tugas_id, kind, name, path, created_at FROM hasil WHERE pekerjaan_id = ? ORDER BY created_at, id`, pekerjaanID)
	if err != nil {
		return nil, fmt.Errorf("daftar hasil: %w", err)
	}
	defer rows.Close()

	var result []domain.Hasil
	for rows.Next() {
		var hasil domain.Hasil
		var tugasID sql.NullString
		var created string
		if err := rows.Scan(&hasil.ID, &hasil.PekerjaanID, &tugasID, &hasil.Kind, &hasil.Name, &hasil.Path, &created); err != nil {
			return nil, fmt.Errorf("baca hasil: %w", err)
		}
		if tugasID.Valid {
			value := domain.ID(tugasID.String)
			hasil.TugasID = &value
		}
		var parseErr error
		if hasil.CreatedAt, parseErr = parseTime(created); parseErr != nil {
			return nil, parseErr
		}
		result = append(result, hasil)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("baca daftar hasil: %w", err)
	}
	return result, nil
}

func validateHasil(hasil domain.Hasil) error {
	if strings.TrimSpace(string(hasil.ID)) == "" || strings.TrimSpace(string(hasil.PekerjaanID)) == "" || strings.TrimSpace(hasil.Kind) == "" || strings.TrimSpace(hasil.Name) == "" || strings.TrimSpace(hasil.Path) == "" {
		return ErrInvalid
	}
	if strings.ContainsRune(hasil.Name, '\x00') || strings.ContainsAny(hasil.Name, `/\\`) {
		return ErrInvalid
	}
	if strings.ContainsRune(hasil.Path, '\x00') || strings.ContainsRune(hasil.Path, '\\') || path.IsAbs(hasil.Path) {
		return ErrInvalid
	}
	clean := path.Clean(hasil.Path)
	if clean != hasil.Path || clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return ErrInvalid
	}
	return nil
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
