package storage

import (
	"strings"
	"testing"
)

func TestResultStorePutRead(t *testing.T) {
	store, err := NewResultStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewResultStore() error = %v", err)
	}
	if err := store.Put("pekerjaan-1/hasil.txt", strings.NewReader("halo Klip")); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	got, err := store.Read("pekerjaan-1/hasil.txt")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}
	if string(got) != "halo Klip" {
		t.Fatalf("Read() = %q, want %q", got, "halo Klip")
	}
}

func TestResultStoreRejectsUnsafePath(t *testing.T) {
	store, err := NewResultStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewResultStore() error = %v", err)
	}
	for _, path := range []string{"../rahasia.txt", "/tmp/hasil.txt", "", "./hasil.txt"} {
		t.Run(path, func(t *testing.T) {
			if err := store.Put(path, strings.NewReader("x")); err == nil {
				t.Fatal("Put() error = nil, want error")
			}
		})
	}
}

func TestResultStoreRejectsOversize(t *testing.T) {
	store, err := NewResultStore(t.TempDir())
	if err != nil {
		t.Fatalf("NewResultStore() error = %v", err)
	}
	reader := strings.NewReader(strings.Repeat("x", int(maxResultSize)+1))
	if err := store.Put("besar.bin", reader); err == nil {
		t.Fatal("Put() error = nil, want size error")
	}
}
