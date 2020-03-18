package main

import (
	"crypto/tls"
	"fmt"
	"golang.org/x/net/http2"
	"net"
	"net/http"
)

func main() {
	const url = "https://d6bc7dd3a64245c0baae023d61cf84d6-cn-hangzhou.alicloudapi.com"
	client := http.Client{
		Transport: &http2.Transport{
			// So http2.Transport doesn't complain the URL scheme isn't 'https'
			AllowHTTP: true,
			// Pretend we are dialing a TLS endpoint.
			// Note, we ignore the passed tls.Config
			DialTLS: func(network, addr string, cfg *tls.Config) (net.Conn, error) {
				return net.Dial(network, addr)
			},
		},
	}
	resp, _ := client.Get(url)
	fmt.Printf("Client Proto: %v\n", resp)
	defer resp.Body.Close()
	var buffer = make([]byte, 1024)
	n, _ := resp.Body.Read(buffer)
	content := string(buffer[:n])
	fmt.Printf("Client content: %s\n", content)

}
