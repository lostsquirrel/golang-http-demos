package sockets

import "errors"

const Socket5Version = 0x05

func CheckVersion(s *Socks5Resolution) error {
	if s.VER != Socket5Version {
		return errors.New("protocol not match")
	}
	return nil
}
