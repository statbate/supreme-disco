package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

type enterChanel struct {
	Conn   *websocket.Conn
	Chanel string
}

type wsMessage struct {
	Message []byte
	Decode  bool
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		if r.Header.Get("origin") == "https://statbate.com" {
			return true
		}
		return false
	},
}

var (
	wsClients = make(map[*websocket.Conn]string)

	json = jsoniter.ConfigCompatibleWithStandardLibrary

	ws = struct {
		Count chan int
		Add   chan *websocket.Conn
		Del   chan *websocket.Conn
		Send  chan wsMessage
		Enter chan enterChanel
	}{
		Count: make(chan int, 100),
		Add:   make(chan *websocket.Conn, 100),
		Del:   make(chan *websocket.Conn, 100),
		Send:  make(chan wsMessage, 100),
		Enter: make(chan enterChanel, 100),
	}
)

func broadcast() {
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case conn := <-ws.Add:
			wsClients[conn] = ""

		case conn := <-ws.Del:
			delete(wsClients, conn)

		case <-ws.Count:
			ws.Count <- len(wsClients)

		case r := <-ws.Enter:
			wsClients[r.Conn] = r.Chanel

		case r := <-ws.Send:
			sendMessage(r.Message, r.Decode)

		case <-ticker.C:
			sendMessage([]byte("ping"), false)
		}
	}
}

func sendMessage(message []byte, decode bool) {
	chanel := ""
	if decode {
		var input map[string]interface{}
		if err := json.Unmarshal(message, &input); err != nil {
			fmt.Println("json error: ", err.Error())
			return
		}
		if input["chanel"] != nil {
			chanel = input["chanel"].(string)
		}
	}

	fmt.Println(string(message), decode, chanel)

	for conn, ch := range wsClients {
		if decode && ch != chanel {
			continue
		}
		if err := conn.WriteMessage(1, message); err != nil {
			conn.Close()
			delete(wsClients, conn)
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	go readWS(conn)
}

func readWS(conn *websocket.Conn) {
	defer conn.Close()

	ws.Add <- conn

	defer func() {
		ws.Del <- conn
	}()

	chanels := map[string]bool{
		"chaturbate": true,
		"bongacams":  true,
		"stripchat":  true,
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("readWS", err.Error())
			return
		}

		input := string(message)

		if len(input) < 32 && chanels[input] {
			ws.Enter <- enterChanel{Conn: conn, Chanel: input}
		}
	}
}
