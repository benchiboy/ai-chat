package main

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

//Group maintains the set of active clients and workers
type Group struct {
	//id to client.
	client_id_conn *sync.Map

	//conn to client
	client_conn_client *sync.Map

	//id to client.
	worker_id_conn *sync.Map

	//conn to client
	worker_conn_client *sync.Map

	//Active Clients count
	client_cnt int

	//Active Clients count
	worker_cnt int

	//total connect count
	total_cnt int

	// Inbound messages from the clients.
	read_user chan *Client

	write_user chan *Client

	// Register requests from the clients.
	client_register chan *Client

	// Unregister requests from clients.
	client_unregister chan *Client

	// Register requests from the clients.
	worker_register chan *Client

	// Unregister requests from clients.
	worker_unregister chan *Client
}

func newGroup() *Group {
	return &Group{

		read_user:  make(chan *Client, 1),
		write_user: make(chan *Client, 10),

		client_cnt: 0,
		worker_cnt: 0,
		total_cnt:  0,

		client_register:   make(chan *Client),
		client_unregister: make(chan *Client),
		worker_register:   make(chan *Client),
		worker_unregister: make(chan *Client),

		client_id_conn:     &sync.Map{},
		client_conn_client: &sync.Map{},

		worker_id_conn:     &sync.Map{},
		worker_conn_client: &sync.Map{},
	}
}

func (h *Group) run() {
	for {
		select {
		case client := <-h.client_register:
			log.Println("Client Register....")
			h.client_id_conn.Store(client.from_user_id, client.conn)
			h.client_conn_client.Store(client.conn, client)
			h.client_cnt++
			h.total_cnt++

		case client := <-h.client_unregister:
			log.Println("Client UnRegister....")
			if _, ok := h.client_id_conn.Load(client.from_user_id); ok {
				h.client_id_conn.Delete(client.from_user_id)
				h.client_conn_client.Delete(client.conn)
				h.client_cnt--
				h.total_cnt--
			} else {
				log.Println("Client UnRegistered")
			}

		case worker := <-h.worker_register:
			log.Println("Worker Register....")
			h.worker_id_conn.Store(worker.from_user_id, worker.conn)
			h.worker_conn_client.Store(worker.conn, worker)
			h.worker_cnt++
			h.total_cnt++

		case worker := <-h.worker_unregister:
			log.Println("Worker UnRegister....")
			if _, ok := h.worker_id_conn.Load(worker.from_user_id); ok {
				h.worker_id_conn.Delete(worker.from_user_id)
				h.worker_conn_client.Delete(worker.conn)
				h.worker_cnt--
				h.total_cnt--
			} else {
				log.Println("Worker UnRegistered")
			}
		case client := <-h.read_user:
			log.Println("is_worker====>", client.is_worker)
			if client.is_worker {
				log.Println("坐席发起的消息")
				if client.to_user_id != "" {
					log.Println("有过交流")
				}
			}
			if !client.is_worker {
				log.Println("客户发起的消息======>")
				h.write_user <- client
				c := h.getClient(client.from_user_id, client.to_user_id)
				if c == nil {
					log.Println("机器人应答消息")
					r := new(Client)
					r.conn = client.conn
					r.group = client.group
					r.send = []byte(h.getRobotMsg(string(client.send)))
					h.write_user <- r
				} else {
					client.to_user_id = c.from_user_id
					c.to_user_id = client.from_user_id
					c.send = []byte(string(client.send) + "hello.....")
					h.write_user <- c
				}
			}
		}
	}
}

/*

 */
func (h *Group) getClient(from_user_id string, to_user_id string) *Client {
	var newClient *Client
	if to_user_id != "" {
		ct, ok := h.worker_id_conn.Load(to_user_id)
		if ok {
			log.Println("得到之前的客服=====>")
			conn, _ := ct.(*websocket.Conn)
			nt, _ := h.worker_conn_client.Load(conn)
			node, _ := nt.(*Client)
			return node
		}
	}
	//首次从客服池子获取
	var bfind bool
	h.worker_conn_client.Range(func(k, v interface{}) bool {
		worker, _ := v.(*Client)
		newClient = worker
		bfind = true
		return false
	})
	if bfind {
		log.Println("从池子获取客服")
		return newClient
	}
	return nil
}

func (h *Group) getRobotMsg(input_msg string) string {

	return input_msg + "hello"

}
