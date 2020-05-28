package main

import (
	"log"
	"sync"
	//	"github.com/gorilla/websocket"
)

//Group maintains the set of active clients and workers
type Group struct {
	//id to client.
	id_conn *sync.Map

	//conn to client
	conn_client *sync.Map

	//Active Clients count
	client_cnt int

	// Inbound messages from the clients.
	from_user chan *Client

	// Inbound messages from the clients.
	to_user chan *Client

	message chan string

	// Register requests from the clients.
	client_register chan *Client

	// Unregister requests from clients.
	client_unregister chan *Client
}

func newGroup() *Group {
	return &Group{
		from_user:  make(chan *Client, 10),
		to_user:    make(chan *Client, 10),
		message:    make(chan string, 1),
		client_cnt: 0,

		client_register:   make(chan *Client),
		client_unregister: make(chan *Client),

		id_conn:     &sync.Map{},
		conn_client: &sync.Map{},
	}
}

func (h *Group) run() {
	for {
		select {
		case client := <-h.client_register:
			h.id_conn.Store(client.from_user_id, client.conn)
			h.conn_client.Store(client.conn, client)
			h.client_cnt++

		case client := <-h.client_unregister:
			if _, ok := h.conn_client.Load(client.conn); ok {
				h.id_conn.Delete(client.from_user_id)
				h.conn_client.Delete(client.conn)
				h.client_cnt--
			}

		case msg := <-h.message:
			
			log.Println("select", msg)
		case client := <-h.from_user:
			log.Println("Recv a Message===>,完成决策=====》", string(client.send))
			h.from_user <- client
			log.Println("from_user", client.from_user_id)

			//			h.id_conn.Range(func(k, v interface{}) bool {
			//				vv, _ := k.(string)
			//				if client.from_user_id != vv {
			//					log.Println("to_user===>", vv)
			//					conn, _ := v.(*websocket.Conn)
			//					n, _ := h.conn_client.Load(conn)
			//					newC, _ := n.(*Client)
			//					newC.send = client.send
			//					h.to_user <- newC
			//					log.Println("to_user ok")
			//				}
			//				return true
			//			})

		}
	}
}

func (h *Group) toWork() {

}

func (h *Group) torRobot() {

}
