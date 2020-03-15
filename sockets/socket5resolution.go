package sockets

import (
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"net"
)

const (
	CONNECT = 0x01
	BIND    = 0x02
	UDP     = 0x03

	IPv4   = 0x01
	Domain = 0x03
	IPv6   = 0x04

	leastMessageLength = 7
)

type Socks5Resolution struct {
	VER         uint8  // socket version
	CMD         uint8  //  socket command
	RSV         uint8  // reserved
	AddressType uint8  // dst address type
	DstAddr     []byte // dst address
	DstPort     uint16 // dst port
	DstDomain   string
	RawAddr     *net.TCPAddr
}

func (s *Socks5Resolution) lstRequest(conn net.Conn) error {
	b := make([]byte, 128)
	n, err := conn.Read(b)
	if err != nil || n < leastMessageLength {
		log.Println(err)
		return errors.New("protocol error")
	}
	s.VER = b[0]
	err2 := CheckVersion(s)
	if err2 != nil {
		return err2
	}
	s.CMD = b[1]
	if s.CMD != CONNECT {
		return errors.New("not supported command")
	}
	s.RSV = b[2] // reserved
	s.AddressType = b[3]
	switch s.AddressType {
	case IPv4:
		//	IP V4 address: X'01'
		s.DstAddr = b[4 : 4+net.IPv4len]
	case Domain:
		//	DOMAINNAME: X'03'
		s.DstDomain = string(b[5 : n-2])
		ipAddr, err := net.ResolveIPAddr("ip", s.DstDomain)
		if err != nil {
			return err
		}
		s.DstAddr = ipAddr.IP
	case IPv6:
		//	IP V6 address: X'04'
		s.DstAddr = b[4 : 4+net.IPv6len]
	default:
		return errors.New(fmt.Sprintf("address error with type %d", s.AddressType))
	}
	s.DstPort = binary.BigEndian.Uint16(b[n-2 : n])
	s.RawAddr = &net.TCPAddr{
		IP:   s.DstAddr,
		Port: int(s.DstPort),
	}

	/**
	回应客户端,响应客户端连接成功
	+----+-----+-------+------+----------+----------+
	|VER | REP | RSV | ATYP | BND.ADDR | BND.PORT |
	+----+-----+-------+------+----------+----------+
	| 1 | 1 | X'00' | 1 | Variable | 2 |
	+----+-----+-------+------+----------+----------+
	*/
	resp := []byte{Socket5Version, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	n, err = conn.Write(resp)
	return err
}
