package custom

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"
	"syscall"
)

var hasPort = regexp.MustCompile(`:\d+$`)

type ConnectAction int

const (
	ConnectProxy = 1
	ConnectMM    = 2
)

func isConnectionClosed(err error) bool {
	if err == nil {
		return false
	}
	if err == io.EOF {
		return true
	}
	i := 0
	var newerr = &err
	for opError, ok := (*newerr).(*net.OpError); ok && i < 10; {
		i++
		newerr = &opError.Err
		if syscallError, ok := (*newerr).(*os.SyscallError); ok {
			if syscallError.Err == syscall.EPIPE || syscallError.Err == syscall.ECONNRESET || syscallError.Err == syscall.EPROTOTYPE {
				return true
			}
		}
	}
	return false
}

type ProxyHandler struct {
	MyRoundTripper http.RoundTripper
	ConnectAction  ConnectAction
}

func (f ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !r.URL.IsAbs() {
		w.WriteHeader(http.StatusBadRequest)
	}
	log.Printf("url %v", r.URL)
	log.Printf("url string %s", r.URL.String())
	log.Print(r)

	if r.Method == "CONNECT" {
		hij, ok := w.(http.Hijacker)
		if !ok {
			if r.Body != nil {
				defer r.Body.Close()
			}
			log.Printf("Connect %s", "hijacking not supported")
			return
		}
		conn, _, err := hij.Hijack()
		if err != nil {
			if r.Body != nil {
				defer r.Body.Close()
			}
			log.Print("Connect", "hijacking not supported")
			return
		}
		hijConn := conn
		host := r.URL.Host
		if !hasPort.MatchString(host) {
			host += ":80"
		}
		if f.ConnectAction == ConnectProxy {
			conn, err := net.Dial("tcp", host)
			if err != nil {
				hijConn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				hijConn.Close()
				log.Print("Connect", "ErrRemoteConnect")
				return
			}
			remoteConn := conn.(*net.TCPConn)
			if _, err := hijConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
				hijConn.Close()
				remoteConn.Close()
				if !isConnectionClosed(err) {
					log.Print("Connect", "ErrResponseWrite")
				}
				return
			}
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				defer func() {
					e := recover()
					err, ok := e.(error)
					if !ok {
						return
					}
					hijConn.Close()
					remoteConn.Close()
					if !isConnectionClosed(err) {
						log.Print("Connect", "ErrRequestRead")
					}
				}()
				_, err := io.Copy(remoteConn, hijConn)
				if err != nil {
					panic(err)
				}
				remoteConn.CloseWrite()
				if c, ok := hijConn.(*net.TCPConn); ok {
					c.CloseRead()
				}
			}()
			go func() {
				defer wg.Done()
				defer func() {
					e := recover()
					err, ok := e.(error)
					if !ok {
						return
					}
					hijConn.Close()
					remoteConn.Close()
					if !isConnectionClosed(err) {
						log.Print("Connect", "ErrResponseWrite")
					}
				}()
				_, err := io.Copy(hijConn, remoteConn)
				if err != nil {
					panic(err)
				}
				remoteConn.CloseRead()
				if c, ok := hijConn.(*net.TCPConn); ok {
					c.CloseWrite()
				}
			}()
			wg.Wait()
			hijConn.Close()
			remoteConn.Close()
		}

		return
	}

	resp, err := f.MyRoundTripper.RoundTrip(r)
	if err != nil {
		log.Print(err)
	}
	resp.Request = r
	h := w.Header()
	for k, v := range resp.Header {
		for _, v1 := range v {
			fmt.Printf("header %s: %s\n", k, v1)
			h.Add(k, v1)
		}
	}
	if resp.Body != nil {
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Print(err)
		}
	}

	//fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func CreateProxyServer() {
	handler := ProxyHandler{
		MyRoundTripper: &http.Transport{TLSClientConfig: &tls.Config{},
			Proxy: http.ProxyFromEnvironment},
		ConnectAction: ConnectProxy,
	}

	http.ListenAndServe(":9000", handler)
}
