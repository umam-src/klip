package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "path/filepath"

    "github.com/umam-src/klip/internal/storage"
)

func main() {
    ctx := context.Background()

    dataDir, err := os.UserConfigDir()
    if err != nil {
        log.Fatalf("gagal menentukan direktori data: %v", err)
    }
    dataDir = filepath.Join(dataDir, "Klip")
    if err := os.MkdirAll(dataDir, 0o700); err != nil {
        log.Fatalf("gagal membuat direktori data: %v", err)
    }

    db, err := storage.Open(ctx, filepath.Join(dataDir, "klip.db"))
    if err != nil {
        log.Fatalf("gagal membuka database: %v", err)
    }
    defer db.Close()

    http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        _, _ = w.Write([]byte("ok\n"))
    })

    fmt.Println("Klip — orkestrator AI lokal")
    log.Printf("data: %s", dataDir)
    log.Println("menjalankan server di http://127.0.0.1:8787")
    log.Fatal(http.ListenAndServe("127.0.0.1:8787", nil))
}
