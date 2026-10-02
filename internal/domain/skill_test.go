package domain

import "testing"

func TestIsValidSkillName(t *testing.T) {
	tests := map[string]bool{
		"ringkas":       true,
		"ringkas-2":     true,
		"":              false,
		"Ringkas":       false,
		"ringkas_2":     false,
		"../ringkas":    false,
		"ringkas/lanjut": false,
		"-ringkas":      false,
		"ringkas-":      false,
	}
	for name, want := range tests {
		t.Run(name, func(t *testing.T) {
			if got := IsValidSkillName(name); got != want {
				t.Fatalf("IsValidSkillName(%q) = %v, want %v", name, got, want)
			}
		})
	}
}
