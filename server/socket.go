package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	
	jsoniter "github.com/json-iterator/go"
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
	d := json.NewDecoder(conn)
	for {
		raw := jsoniter.RawMessage{}
		err := d.Decode(&raw)
		if err == io.EOF {
			fmt.Println("break:", err.Error())
			break
		} else if err != nil {
			fmt.Println("json error:", err.Error())
			continue
		}
		ws.Broadcast <- raw
	}
}
