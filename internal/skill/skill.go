package skill

import (
	"errors"
	"fmt"
	"strings"

	"github.com/umam-src/klip/internal/domain"
)

var (
	ErrInvalidFormat = errors.New("format skill tidak valid")
	ErrMissingField  = errors.New("field skill wajib diisi")
)

// Skill adalah definisi instruksi lokal yang dapat dibaca Klip.
// Format penyimpanan yang didukung adalah SKILL.md dengan front matter
// sederhana berisi name dan description, diikuti isi instruksi Markdown.
type Skill struct {
	Name         string
	Description  string
	Instructions string
}

// Parse membaca satu dokumen SKILL.md tanpa memerlukan parser YAML eksternal.
// Front matter harus berada di antara dua baris "---" pertama.
func Parse(content string) (Skill, error) {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.TrimSpace(content)
	if content == "" {
		return Skill{}, fmt.Errorf("%w: dokumen kosong", ErrInvalidFormat)
	}

	lines := strings.Split(content, "\n")
	if len(lines) < 4 || strings.TrimSpace(lines[0]) != "---" {
		return Skill{}, fmt.Errorf("%w: front matter wajib diawali ---", ErrInvalidFormat)
	}

	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end < 0 {
		return Skill{}, fmt.Errorf("%w: penutup front matter tidak ditemukan", ErrInvalidFormat)
	}

	fields := make(map[string]string, 2)
	for _, line := range lines[1:end] {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			return Skill{}, fmt.Errorf("%w: baris front matter %q", ErrInvalidFormat, line)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key != "name" && key != "description" {
			return Skill{}, fmt.Errorf("%w: field %q tidak didukung", ErrInvalidFormat, key)
		}
		if value == "" {
			return Skill{}, fmt.Errorf("%w: %s", ErrMissingField, key)
		}
		if _, exists := fields[key]; exists {
			return Skill{}, fmt.Errorf("%w: field %q ganda", ErrInvalidFormat, key)
		}
		fields[key] = value
	}

	name := fields["name"]
	if name == "" {
		return Skill{}, fmt.Errorf("%w: name", ErrMissingField)
	}
	if !domain.IsValidSkillName(name) {
		return Skill{}, fmt.Errorf("%w: name harus huruf kecil, angka, dan tanda hubung", ErrInvalidFormat)
	}
	if fields["description"] == "" {
		return Skill{}, fmt.Errorf("%w: description", ErrMissingField)
	}

	instructions := strings.TrimSpace(strings.Join(lines[end+1:], "\n"))
	if instructions == "" {
		return Skill{}, fmt.Errorf("%w: instructions", ErrMissingField)
	}

	return Skill{
		Name:         name,
		Description:  fields["description"],
		Instructions: instructions,
	}, nil
}
