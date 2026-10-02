package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})

	fmt.Println("Klip — orkestrator AI lokal")
	log.Println("menjalankan server di http://127.0.0.1:8787")
	log.Fatal(http.ListenAndServe("127.0.0.1:8787", nil))
}
