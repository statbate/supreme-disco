package main

import (
	"net"
	"time"
)

var echoSocket = make(chan []byte)

func startSocket() {
	for {
		select {
		case b := <-echoSocket:
			conn, err := net.Dial("unix", "/tmp/echo.sock")
			if err == nil {
				conn.Write(b)
				conn.Close()
			}
		}
	}
}

func main() {
	go startSocket()

	for {
		echoSocket <- []byte(`{"chanel":"stripchat","room":"naughtyobsessions","donator":"tbone6893","amount":333}`)
		echoSocket <- []byte(`{"chanel":"stripchat","index":28}`)
		time.Sleep(1e9)
	}
}
