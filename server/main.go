package main

import (
	"log"
	"net"
	"net/http"
	"os"
)

func main() {

	go broadcast()
	go startSocket()

	http.HandleFunc("/ws/", wsHandler)

	const SOCK = "/tmp/ws.sock"
	removeSocket(SOCK)
	unixListener, err := net.Listen("unix", SOCK)
	if err != nil {
		log.Fatal("Listen (ws socket): ", err)
	}
	defer unixListener.Close()
	os.Chmod(SOCK, 0777)
	log.Fatal(http.Serve(unixListener, nil))
}
