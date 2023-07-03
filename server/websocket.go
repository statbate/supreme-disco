package main

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
)

type Client struct {
	Chanel string
	Conn   *websocket.Conn
}

var (
	wsClients = make(map[*Client]bool)

	json = jsoniter.ConfigCompatibleWithStandardLibrary

	ws = struct {
		Broadcast chan []byte
		Add       chan *Client
		Del       chan *Client
	}{
		Broadcast: make(chan []byte, 1),
		Add:       make(chan *Client),
		Del:       make(chan *Client),
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
	ticker := time.NewTicker(20 * time.Second)
	for {
		select {
		case client := <-ws.Add:
			wsClients[client] = true

		case conn := <-ws.Del:
			delete(wsClients, conn)

		case r := <-ws.Broadcast:
			//fmt.Println(len(ws.Broadcast), cap(ws.Broadcast))
			sendBroadcast(r)

		case <-ticker.C:
			fmt.Println(len(wsClients), runtime.NumGoroutine(), len(ws.Broadcast), cap(ws.Broadcast))
		}
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
		if err := client.Conn.WriteMessage(1, message); err != nil {
			delete(wsClients, client)
			client.Conn.Close()
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

	client := &Client{Conn: conn, Chanel: input.Chanel}

	chanels := map[string]bool{
		"chaturbate": true,
		"bongacams":  true,
		"stripchat":  true,
		"camsoda":    true,
	}

	if !chanels[input.Chanel] {
		return
	}

	ws.Add <- client

	defer func() {
		ws.Del <- client
	}()

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Minute))
		_, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
	}
}
