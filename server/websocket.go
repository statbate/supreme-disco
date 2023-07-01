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

type sendMsg struct {
	Conn   *websocket.Conn
	Message []byte
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
		Broadcast  chan []byte
		Send  chan sendMsg
		Enter chan enterChanel
		Add   chan *websocket.Conn
		Del   chan *websocket.Conn
	}{
		Broadcast:  make(chan []byte, 100),
		Send:  make(chan sendMsg, 100),
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
			//fmt.Println("add conn", len(wsClients))

		case conn := <-ws.Del:
			delete(wsClients, conn)
			//fmt.Println("delete conn", len(wsClients))

		case r := <-ws.Enter:
			wsClients[r.Conn] = r.Chanel

		case r := <-ws.Send:
			sendMessage(r.Conn, r.Message)
			
		case r := <-ws.Broadcast:
			sendBroadcast(r)

		case <-ticker.C:
			fmt.Println("WebSocket:", len(wsClients))
		}
	}
}

func sendMessage(conn *websocket.Conn, message []byte){
	if err := conn.WriteMessage(1, message); err != nil {
		conn.Close()
	}
}

func sendBroadcast(message []byte) {
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
		sendMessage(conn, message)
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
			//fmt.Println("readWS", err.Error())
			return
		}

		if string(message) == "ping" {
			if time.Now().Unix() > ping {
				ws.Send <- sendMsg{Conn: conn, Message: []byte("pong")}
				ping = time.Now().Unix() + 15
			}
			continue
		}

		input := struct {
			Chanel string `json:"chanel"`
		}{}

		if err := json.Unmarshal(message, &input); err != nil {
			//fmt.Println("wrong json", err.Error())
			return
		}

		if chanels[input.Chanel] {
			ws.Enter <- enterChanel{Conn: conn, Chanel: input.Chanel}
		}
	}
}
