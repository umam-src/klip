package i18n

import "testing"

func TestNormalize(t *testing.T) {
	if got := Normalize(DefaultLocale); got != DefaultLocale {
		t.Fatalf("Normalize(default) = %q, want %q", got, DefaultLocale)
	}
	if got := Normalize("en-US"); got != DefaultLocale {
		t.Fatalf("Normalize(unsupported) = %q, want %q", got, DefaultLocale)
	}
}

func TestSupported(t *testing.T) {
	got := Supported()
	if len(got) != 1 || got[0] != DefaultLocale {
		t.Fatalf("Supported() = %#v, want [%q]", got, DefaultLocale)
	}
}
