package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

type Client struct {
	Chanel string
	Conn   *websocket.Conn
}

type sendMsg struct {
	Message []byte
	Conn    *websocket.Conn
}

var (
	wsClients = make(map[*Client]bool)

	json = jsoniter.ConfigCompatibleWithStandardLibrary

	ws = struct {
		Broadcast chan []byte
		Send      chan sendMsg
		Add       chan *Client
		Del       chan *Client
	}{
		Broadcast: make(chan []byte, 100),
		Send:      make(chan sendMsg, 100),
		Add:       make(chan *Client, 100),
		Del:       make(chan *Client, 100),
	}

	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			params := r.URL.Query()
			if params["api_key"] != nil || r.Header.Get("origin") == "https://statbate.com" {
				return true
			}
			return false
		},
	}
)

func broadcast() {
	ticker := time.NewTicker(30 * time.Second)
	for {
		select {
		case client := <-ws.Add:
			wsClients[client] = true

		case conn := <-ws.Del:
			delete(wsClients, conn)

		case r := <-ws.Send:
			sendMessage(r.Conn, r.Message)

		case r := <-ws.Broadcast:
			sendBroadcast(r)

		case <-ticker.C:
			fmt.Println("WebSocket:", len(wsClients))
		}
	}
}

func sendMessage(conn *websocket.Conn, message []byte) {
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
	for client := range wsClients {
		if client.Chanel != input.Chanel {
			continue
		}
		sendMessage(client.Conn, message)
	}
}

func enterChannel(conn *websocket.Conn, chanel string) (*Client, bool) {
	chanels := map[string]bool{
		"chaturbate": true,
		"bongacams":  true,
		"stripchat":  true,
		"camsoda":    true,
	}
	client := &Client{Conn: conn, Chanel: chanel}
	if chanels[chanel] {
		ws.Add <- client
		return client, true
	}
	return client, false
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

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		return
	}

	input := struct {
		Chanel string `json:"chanel"`
	}{}

	if err := json.Unmarshal(message, &input); err != nil {
		return
	}

	client, ok := enterChannel(conn, input.Chanel)
	if !ok {
		return
	}

	defer func() {
		ws.Del <- client
	}()

	ping := time.Now().Unix()
	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Minute))
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		if string(message) == "ping" {
			if time.Now().Unix() > ping {
				ws.Send <- sendMsg{Conn: conn, Message: []byte("pong")}
				ping = time.Now().Unix() + 15
			}
			continue
		}
	}
}
