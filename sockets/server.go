package sockets

import (
	"io"
	"log"
	"net"
	"sync"
)

func handleClientRequest(client *net.TCPConn) {
	if client == nil {
		return
	}
	defer client.Close()

	// create buffer
	//buff := make([]byte, 255)

	// 认证协商
	var proto Socket5
	err := proto.handshake(client)
	if err != nil {
		log.Print(client.RemoteAddr(), err)
		return
	}

	var request Socks5Resolution

	err = request.lstRequest(client)
	if err != nil {
		log.Print(client.RemoteAddr(), err)
		return
	}

	log.Println(client.RemoteAddr(), request.DstDomain, request.DstAddr, request.DstPort)

	//
	dstServer, err := net.DialTCP("tcp", nil, request.RawAddr)
	if err != nil {
		log.Print(client.RemoteAddr(), err)
		return
	}
	defer dstServer.Close()

	wg := new(sync.WaitGroup)
	wg.Add(2)

	// 本地的内容copy到远程端
	go func() {
		defer wg.Done()
		io.Copy(client, dstServer)
	}()

	// 远程得到的内容copy到源地址
	go func() {
		defer wg.Done()
		io.Copy(dstServer, client)
	}()
	wg.Wait()

}

func Server(listenAddrString string) {

	// listen on the address
	listenAddr, err := net.ResolveTCPAddr("tcp", listenAddrString)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("listen on the address: %s ", listenAddrString)

	listener, err := net.ListenTCP("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Fatal(err)
		}
		go handleClientRequest(conn)
	}
}
