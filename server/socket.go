package main

import (
	"log"
	"net"
	"os"
)

func removeSocket(s string) {
	if _, err := os.Stat(s); err == nil {
		os.Remove(s)
	}
}

func startSocket() {
	const SOCK = "/tmp/echo.sock"
	removeSocket(SOCK)
	listener, err := net.Listen("unix", SOCK)
	if err != nil {
		log.Fatal("Listen (echo socket): ", err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("accept error:", err)
		}
		go socketHandler(conn)
	}
}

func socketHandler(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 512)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			return
		}
		ws.Send <- wsMessage{Message: buf[0:n], Decode: true}
	}
}
