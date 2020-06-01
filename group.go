package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

//Group maintains the set of active clients and workers
type Group struct {
	//group_id
	group_id string

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

func newGroup(group_id string) *Group {
	return &Group{
		group_id: group_id,

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
			PrintInfo("Client Register....")
			h.client_id_conn.Store(client.from_user_id, client.conn)
			h.client_conn_client.Store(client.conn, client)
			h.client_cnt++
			h.total_cnt++
			h.worker_conn_client.Range(func(k, v interface{}) bool {
				client, _ := v.(*Client)
				client.send = h.getClients()
				h.write_user <- client
				return true
			})
		case client := <-h.client_unregister:
			PrintInfo("Client UnRegister....")
			if _, ok := h.client_id_conn.Load(client.from_user_id); ok {
				h.client_id_conn.Delete(client.from_user_id)
				h.client_conn_client.Delete(client.conn)
				h.client_cnt--
				h.total_cnt--
				h.worker_conn_client.Range(func(k, v interface{}) bool {
					client, _ := v.(*Client)
					client.send = h.getClients()
					h.write_user <- client
					return true
				})
			}
		case worker := <-h.worker_register:
			PrintInfo("Worker Register....")
			h.worker_id_conn.Store(worker.from_user_id, worker.conn)
			h.worker_conn_client.Store(worker.conn, worker)
			h.worker_cnt++
			h.total_cnt++

		case worker := <-h.worker_unregister:
			PrintInfo("Worker UnRegister....")
			if _, ok := h.worker_id_conn.Load(worker.from_user_id); ok {
				h.worker_id_conn.Delete(worker.from_user_id)
				h.worker_conn_client.Delete(worker.conn)
				h.worker_cnt--
				h.total_cnt--
			}
		case client := <-h.read_user:
			if client.is_worker {
				h.doWorker(client, client.send)
			} else {
				h.doClient(client, client.send)
			}
		}
	}
}

/*
	客服的处理
*/
func (h *Group) doWorker(client *Client, message []byte) {
	PrintHead("DoWorker")
	var text MsgText
	json.Unmarshal([]byte(string(client.send)), &text)
	switch text.MsgType {
	case MSG_TEXT:
		h.write_user <- client
		c := h.getClientById(text.ToUserId)
		if c == nil {
			PrintInfo("必须选择一个用户--->")
			h.write_user <- client
		} else {
			c.send = client.send
			h.write_user <- c
		}
	case MSG_USER_LIST:
		client.send = h.getClients()
		h.write_user <- client
	}
	PrintTail("DoWorker")
}

/*
  访客的处理
*/
func (h *Group) doClient(client *Client, message []byte) {
	PrintHead("DoClient")
	var header MsgHeader
	json.Unmarshal(message, &header)
	switch header.MsgType {
	case MSG_USER_ID:
		PrintInfo("获取用户Id")
		var getUser MsgGetUserId
		getUser.FromUserId = client.from_user_id
		getUser.ToUserId = client.from_user_id
		getUser.MsgType = header.MsgType
		msgBuf, _ := json.Marshal(getUser)
		client.send = msgBuf
		h.write_user <- client
	case MSG_TEXT:
		h.write_user <- client
		var text MsgText
		json.Unmarshal([]byte(string(client.send)), &text)
		worker := h.getWorker(text.ToUserId)
		if worker != nil {
			text.ToUserId = worker.from_user_id
			msgBuf, _ := json.Marshal(text)
			worker.send = msgBuf
			h.write_user <- worker
		} else {
			r := new(Client)
			r.conn = client.conn
			r.group = client.group
			r.send = []byte("")
			h.write_user <- r
		}
	}
	PrintTail("DoClient")
}

/*
  得到一个客服节点信息，如果没有在线的有机器来处理
*/
func (h *Group) getWorker(to_user_id string) *Client {
	var newClient *Client
	if to_user_id != "" {
		ct, ok := h.worker_id_conn.Load(to_user_id)
		if ok {
			conn, _ := ct.(*websocket.Conn)
			nt, _ := h.worker_conn_client.Load(conn)
			node, _ := nt.(*Client)
			return node
		}
	}
	var bfind bool
	h.worker_conn_client.Range(func(k, v interface{}) bool {
		worker, _ := v.(*Client)
		newClient = worker
		bfind = true
		return false
	})
	if bfind {
		return newClient
	}
	//交给机器处理
	return nil
}

/*
	根据访客的ID得到访客节点信息
*/
func (h *Group) getClientById(to_user_id string) *Client {
	ct, ok := h.client_id_conn.Load(to_user_id)
	if ok {
		conn, _ := ct.(*websocket.Conn)
		nt, ok := h.client_conn_client.Load(conn)
		node, ok := nt.(*Client)
		if !ok {
			return nil
		}
		return node
	}
	return nil
}

/*
	得到访问用户的列表
*/
func (h *Group) getClients() []byte {
	var users MsgGetUsers
	t := 0
	h.client_conn_client.Range(func(k, v interface{}) bool {
		client, _ := v.(*Client)
		t++
		var u User
		u.UserId = client.from_user_id
		u.UserName = "访客" + fmt.Sprintf("%d", t)
		users.Users = append(users.Users, u)
		return true
	})
	users.MsgType = MSG_USER_LIST
	msgBuf, _ := json.Marshal(users)
	return msgBuf
}
