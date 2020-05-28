package main

import (
	"log"
	"net/http"

	"github.com/satori/go.uuid"
)

// serveWs handles websocket requests from the peer.
func serveWs(hub *Group, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	u1 := uuid.NewV4()
	log.Println("Has a new Connection...")
	client := &Client{group: hub, conn: conn, from_user_id: u1.String(), send: make([]byte, 0)}
	client.group.client_register <- client
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
