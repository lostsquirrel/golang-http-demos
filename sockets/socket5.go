package sockets

import (
	"errors"
	"log"
	"net"
)

type Method int

type Socket5 struct {
	VER           uint8
	MethodsLength uint8
	METHODS       []uint8
}

const (
	NoAuth              = 0x00
	GSSAPI              = 0x01
	UsernameAndPassword = 0x02
	Unacceptable        = 0xFF
)

func (s *Socket5) handshake(conn net.Conn) error {
	b := make([]byte, 255)
	n, err := conn.Read(b)
	if err != nil {
		log.Println(err)
		return err
	}
	s.VER = b[0] //get the version

	if s.VER != Socket5Version {
		return errors.New("protocol error")
	}
	s.MethodsLength = b[1] //get the methods length
	if n != int(2+s.MethodsLength) {
		return errors.New("protocol not match")
	}
	s.METHODS = b[2 : 2+s.MethodsLength] //read the methods

	// handshake response
	resp := []byte{Socket5Version, NoAuth}
	_, err = conn.Write(resp)
	return err
}
