package skill

import (
	"errors"
	"testing"
)

func TestParse(t *testing.T) {
	got, err := Parse("---\nname: ringkas\ndescription: Merangkum teks\n---\n\nBuat ringkasan singkat.")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if got.Name != "ringkas" {
		t.Fatalf("Name = %q, want ringkas", got.Name)
	}
	if got.Description != "Merangkum teks" {
		t.Fatalf("Description = %q", got.Description)
	}
	if got.Instructions != "Buat ringkasan singkat." {
		t.Fatalf("Instructions = %q", got.Instructions)
	}
}

func TestParseRejectsInvalidDocument(t *testing.T) {
	tests := map[string]string{
		"tanpa front matter": "name: ringkas\ndescription: x\n---\nisi",
		"nama tidak valid":   "---\nname: Ringkas\ndescription: x\n---\nisi",
		"instruksi kosong":   "---\nname: ringkas\ndescription: x\n---\n",
		"field tidak dikenal": "---\nname: ringkas\nowner: umam\ndescription: x\n---\nisi",
	}

	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := Parse(content)
			if err == nil {
				t.Fatal("Parse() error = nil, want error")
			}
			if !errors.Is(err, ErrInvalidFormat) && !errors.Is(err, ErrMissingField) {
				t.Fatalf("Parse() error = %v, want format or missing-field error", err)
			}
		})
	}
}
