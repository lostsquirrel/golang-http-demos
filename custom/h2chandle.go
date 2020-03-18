package custom

import (
	"fmt"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"net/http"
)

func CreateHTTP2Serve() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello world")
	})

	h2s := &http2.Server{

	}
	h1s := &http.Server{
		Addr: ":9000",
		Handler: h2c.NewHandler(handler, h2s),
	}

	h1s.ListenAndServe()

}
