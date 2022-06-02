package main

import (
	"io"
	"log"
	"net/http"
)

func healthz(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "OK")
}

func main() {
	http.HandleFunc("/healthz", healthz)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
