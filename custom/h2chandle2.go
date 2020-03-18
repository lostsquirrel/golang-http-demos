package custom

import (
	"fmt"
	"golang.org/x/net/http2"
	"net"
	"net/http"
)

func CreateHTTP2Serve2() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Hello world")
	})

	h2s := &http2.Server{}
	l, err := net.Listen("tcp", "0.0.0.0:9000")
	if err != nil {
		fmt.Println(err)
	}
	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err)
		}
		h2s.ServeConn(conn, &http2.ServeConnOpts{
			Handler: handler,
		})
	}

}
