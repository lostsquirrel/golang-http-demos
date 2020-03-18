package custom

import (
	"fmt"
	"html"
	"log"
	"net/http"
	"time"
)

type FooHandler struct{}

func (f FooHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func CreateServe() {
	myHandler := FooHandler{}
	s := &http.Server{
		Addr:           ":9000",
		Handler:        myHandler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(s.ListenAndServe())
}
