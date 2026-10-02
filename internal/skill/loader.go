package skill

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

const MaxSkillFileSize = 1 << 20

var (
	ErrSkillTooLarge = errors.New("ukuran skill terlalu besar")
	ErrDuplicateName = errors.New("nama skill ganda")
)

// LoadDir membaca skill lokal dari subdirektori langsung root.
// Setiap skill harus memiliki file bernama SKILL.md. Direktori dan file
// diurutkan secara deterministik agar hasil tidak bergantung pada filesystem.
func LoadDir(root string) ([]Skill, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("skill: buka direktori: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	skills := make([]Skill, 0, len(entries))
	seen := make(map[string]struct{}, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		path := filepath.Join(root, entry.Name(), "SKILL.md")
		info, err := os.Stat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return nil, fmt.Errorf("skill %q: periksa SKILL.md: %w", entry.Name(), err)
		}
		if !info.Mode().IsRegular() {
			return nil, fmt.Errorf("skill %q: SKILL.md bukan berkas biasa", entry.Name())
		}
		if info.Size() > MaxSkillFileSize {
			return nil, fmt.Errorf("skill %q: %w: %d byte", entry.Name(), ErrSkillTooLarge, info.Size())
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("skill %q: baca SKILL.md: %w", entry.Name(), err)
		}
		parsed, err := Parse(string(content))
		if err != nil {
			return nil, fmt.Errorf("skill %q: parse SKILL.md: %w", entry.Name(), err)
		}
		if _, exists := seen[parsed.Name]; exists {
			return nil, fmt.Errorf("%w: %s", ErrDuplicateName, parsed.Name)
		}
		seen[parsed.Name] = struct{}{}
		skills = append(skills, parsed)
	}

	return skills, nil
}
