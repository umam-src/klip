package web

import "testing"

func TestFilesContainUI(t *testing.T) {
	for _, name := range []string{"index.html", "assets/app.css", "assets/app.js"} {
		if _, err := Files.ReadFile(name); err != nil {
			t.Fatalf("embedded file %q: %v", name, err)
		}
	}
}
