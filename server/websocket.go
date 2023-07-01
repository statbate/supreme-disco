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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		//if r.Header.Get("origin") == "https://statbate.com" {
		//	return true
		//}
		//return false
		return true
	},
}

var (
	wsClients = make(map[*websocket.Conn]string)

	json = jsoniter.ConfigCompatibleWithStandardLibrary

	ws = struct {
		Send  chan []byte
		Enter chan enterChanel
		Add   chan *websocket.Conn
		Del   chan *websocket.Conn
	}{
		Send:  make(chan []byte, 100),
		Enter: make(chan enterChanel, 100),
		Add:   make(chan *websocket.Conn, 100),
		Del:   make(chan *websocket.Conn, 100),
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

		case r := <-ws.Enter:
			wsClients[r.Conn] = r.Chanel

		case r := <-ws.Send:
			sendMessage(r)

		case <-ticker.C:
			fmt.Println("WebSocket:", len(wsClients))
		}
	}
}

func sendMessage(message []byte) {
	input := struct {
		Chanel string `json:"chanel"`
	}{}
	if err := json.Unmarshal(message, &input); err != nil {
		fmt.Println("json error: ", err.Error())
		return
	}
	for conn, ch := range wsClients {
		if ch != input.Chanel {
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
		"camsoda":    true,
	}
	ping := time.Now().Unix()
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("readWS", err.Error())
			return
		}

		if string(message) == "ping" {
			if time.Now().Unix() > ping {
				if err := conn.WriteMessage(1, []byte("pong")); err != nil {
					fmt.Println("write message", err.Error())
					return
				}
				ping = time.Now().Unix() + 15
			}
			continue
		}

		input := struct {
			Chanel string `json:"chanel"`
		}{}

		if err := json.Unmarshal(message, &input); err != nil {
			fmt.Println("wrong json", err.Error())
			return
		}

		if chanels[input.Chanel] {
			ws.Enter <- enterChanel{Conn: conn, Chanel: input.Chanel}
		}
	}
}
