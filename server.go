package main

import (
	"./sockets"
)

func main() {
	sockets.Server(":9009")
}
