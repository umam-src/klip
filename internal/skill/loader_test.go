package skill

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDir(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "z-last", "zeta")
	writeSkill(t, root, "a-first", "alfa")
	if err := os.WriteFile(filepath.Join(root, "README.md"), []byte("abaikan"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "tanpa-skill"), 0o700); err != nil {
		t.Fatal(err)
	}

	got, err := LoadDir(root)
	if err != nil {
		t.Fatalf("LoadDir() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("LoadDir() returned %d skills, want 2", len(got))
	}
	if got[0].Name != "alfa" || got[1].Name != "zeta" {
		t.Fatalf("LoadDir() order = [%s %s], want [alfa zeta]", got[0].Name, got[1].Name)
	}
}

func TestLoadDirRejectsDuplicateNames(t *testing.T) {
	root := t.TempDir()
	writeSkill(t, root, "satu", "sama")
	writeSkill(t, root, "dua", "sama")

	_, err := LoadDir(root)
	if !errors.Is(err, ErrDuplicateName) {
		t.Fatalf("LoadDir() error = %v, want ErrDuplicateName", err)
	}
}

func TestLoadDirRejectsOversizedSkill(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "besar")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	content := make([]byte, MaxSkillFileSize+1)
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), content, 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := LoadDir(root)
	if !errors.Is(err, ErrSkillTooLarge) {
		t.Fatalf("LoadDir() error = %v, want ErrSkillTooLarge", err)
	}
}

func writeSkill(t *testing.T, root, dir, name string) {
	t.Helper()
	path := filepath.Join(root, dir)
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: " + name + "\ndescription: Skill lokal\n---\n\nInstruksi.\n"
	if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
