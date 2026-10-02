package web

import (
    "io/fs"
    "net/http"
    "strings"
)

// Handler serves the embedded local web interface.
func Handler() http.Handler {
    assets, err := fs.Sub(Files, "assets")
    if err != nil {
        panic(err)
    }

    mux := http.NewServeMux()
    mux.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assets))))
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            http.NotFound(w, r)
            return
        }
        data, err := Files.ReadFile("index.html")
        if err != nil {
            http.Error(w, "antarmuka tidak tersedia", http.StatusInternalServerError)
            return
        }
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Header().Set("Cache-Control", "no-cache")
        _, _ = w.Write(data)
    })
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if strings.HasPrefix(r.URL.Path, "/assets/") || r.URL.Path == "/" {
            mux.ServeHTTP(w, r)
            return
        }
        http.NotFound(w, r)
    })
}
