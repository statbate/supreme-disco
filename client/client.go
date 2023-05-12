package main

import (
	"net"
	"time"
	"fmt"
)

var echoSocket = make(chan []byte)

func startSocket() {
	
	var( 
		err error
		conn net.Conn
	)
	
	for {
		select {
		case b := <-echoSocket:
			
			if conn == nil {
				conn, err = net.Dial("unix", "/tmp/echo.sock")
				if err != nil {
					fmt.Println(err.Error())
					continue
				}
			}
			
			if conn != nil {
				if _, err = conn.Write(b); err != nil {
					fmt.Println(err.Error())
					conn.Close()
					conn = nil
				}
			}
			
		}
	}
}

func main() {
	go startSocket()

	for {
		echoSocket <- []byte(`{"chanel":"chaturbate","room":"naughtyobsessions","donator":"tbone6893","amount":333}`)
		echoSocket <- []byte(`{"chanel":"chaturbate","index":28}`)
		time.Sleep(1e9)
	}
}
