package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/satori/go.uuid"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	lock   sync.Mutex
	groups = new(sync.Map)
)

const (
	WORKER = "1"
	CLIENT = "2"
)

// serveWs handles websocket requests from the peer.
func serveWs(w http.ResponseWriter, r *http.Request) {
	PrintHead("Server Websocket")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	vars := r.URL.Query()
	is_worker := vars["is_worker"][0]
	group_id := vars["group_id"][0]

	PrintInfo("group_id=>", group_id, "is_worker=>", is_worker)

	var ngroup *Group
	group, ok := groups.Load(group_id)
	if !ok {
		group := newGroup(group_id)
		groups.Store(group_id, group)
		go group.run()
		ngroup = group
	} else {
		ngroup, _ = group.(*Group)
	}
	u1 := uuid.NewV4()
	client := &Client{group: ngroup, conn: conn, from_user_id: u1.String(), send: make([]byte, 0)}
	if is_worker == WORKER {
		client.is_worker = true
		client.group.worker_register <- client
	} else {
		client.group.client_register <- client
	}
	go client.writePump()
	go client.readPump()

	PrintTail("Server Websocket")
}

func main() {

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(w, r)
	})
	log.Println("Start....")
	http.ListenAndServe("0.0.0.0:7777", nil)
}
