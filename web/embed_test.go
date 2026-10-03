package web

import (
	"io/fs"
	"regexp"
	"strings"
	"testing"
)

func TestFilesContainUI(t *testing.T) {
	for _, name := range []string{"index.html", "assets/app.css", "assets/app.js"} {
		if _, err := Files.ReadFile(name); err != nil {
			t.Fatalf("embedded file %q: %v", name, err)
		}
	}
}

// Setiap skrip di assets harus dimuat oleh index.html, dan setiap skrip yang
// dimuat harus ada. Skrip yang tidak dimuat membuat halamannya tidak berfungsi
// tanpa pesan kesalahan apa pun.
func TestSetiapSkripDimuatIndex(t *testing.T) {
	index, err := Files.ReadFile("index.html")
	if err != nil {
		t.Fatalf("baca index.html: %v", err)
	}

	dimuat := map[string]bool{}
	pola := regexp.MustCompile(`src="/assets/([^"]+\.js)"`)
	for _, cocok := range pola.FindAllStringSubmatch(string(index), -1) {
		dimuat[cocok[1]] = true
		if _, err := Files.ReadFile("assets/" + cocok[1]); err != nil {
			t.Errorf("skrip %q dimuat index.html tetapi tidak ada: %v", cocok[1], err)
		}
	}

	entri, err := fs.ReadDir(Files, "assets")
	if err != nil {
		t.Fatalf("baca direktori assets: %v", err)
	}
	for _, e := range entri {
		if strings.HasSuffix(e.Name(), ".js") && !dimuat[e.Name()] {
			t.Errorf("skrip %q ada di assets tetapi tidak dimuat index.html", e.Name())
		}
	}
}

// Elemen yang dibutuhkan skrip harus ada di index.html.
func TestIndexMemuatElemenYangDipakaiSkrip(t *testing.T) {
	index, err := Files.ReadFile("index.html")
	if err != nil {
		t.Fatalf("baca index.html: %v", err)
	}
	html := string(index)

	kebutuhan := map[string][]string{
		"agent-detail.js": {"workspace", "agent-detail", "agent-title", "agent-detail-card", "back-from-agent", "project-detail", "task-detail"},
		"provider.js":     {"provider-status", "refresh-provider"},
		"settings.js": {
			"settings-list", "settings-provider", "settings-base-url", "settings-model",
			"settings-message", "refresh-settings", "check-provider", "load-models", "save-settings",
		},
	}
	for skrip, daftarID := range kebutuhan {
		for _, id := range daftarID {
			if !strings.Contains(html, `id="`+id+`"`) {
				t.Errorf("%s membutuhkan elemen #%s, tetapi tidak ada di index.html", skrip, id)
			}
		}
	}
}
