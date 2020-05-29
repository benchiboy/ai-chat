package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

var upgrader = websocket.Upgrader{
	// 允许所有CORS跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Group, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	vars := r.URL.Query()
	is_worker := vars["is_worker"][0]

	u1, _ := uuid.NewV4()
	log.Println("Has a new Connection...")
	client := &Client{group: hub, conn: conn, from_user_id: u1.String(), send: make([]byte, 0)}
	if is_worker == "1" {
		client.is_worker = true
		client.group.worker_register <- client
	} else {
		client.group.client_register <- client
	}
	// new goroutines.
	go client.writePump()
	go client.readPump()
}

func main() {

	hub := newGroup()
	go hub.run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	log.Println("Start....")
	http.ListenAndServe("0.0.0.0:7777", nil)
}
