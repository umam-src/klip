package domain

import "regexp"

var skillNamePattern = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// AgenSkill menghubungkan satu agen dengan skill lokal berdasarkan nama skill.
// Nama skill mengikuti format yang sama dengan SKILL.md.
type AgenSkill struct {
	AgenID    ID
	SkillName string
}

// IsValidSkillName memeriksa nama skill yang aman dan deterministik untuk
// digunakan sebagai identitas skill lokal.
func IsValidSkillName(name string) bool {
	return skillNamePattern.MatchString(name)
}
