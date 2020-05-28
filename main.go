package main

import (
	"bytes"
	"log"
	"net/http"
	"sync"
	"time"

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

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

//client node define
type Client struct {
	//client conn
	conn *websocket.Conn
	//client id
	from_user_id string
	//to_id
	to_user_id string
	//last time
	active_time time.Time
	//group id
	group_id string
	//is_worker
	is_worker bool
	//send buffer
	send []byte
	//group
	group *Group
}

//Group maintains the set of active clients and workers
type Group struct {
	//id to client.
	id_conn *sync.Map

	//conn to client
	conn_client *sync.Map

	//Active Clients count
	client_cnt int

	// Inbound messages from the clients.
	broadcast chan *Client

	// Register requests from the clients.
	client_register chan *Client

	// Unregister requests from clients.
	client_unregister chan *Client
}

func newGroup() *Group {
	return &Group{
		broadcast:  make(chan *Client),
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
		case client := <-h.broadcast:
			log.Println("Recv a Message===>,完成决策=====》", string(client.send))
			h.broadcast <- client
		}
	}
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		log.Println("Disconnect.....")
		c.group.client_unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		log.Println("Pong Hander...")
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			log.Println("readPump--->", err.Error())
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.send = message
		c.group.broadcast <- c
	}
}

//writePump pumps messages from the hub to the websocket connection.

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		c.group.client_unregister <- c
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		log.Println("Write Dumpling....")
		select {
		case client := <-c.group.broadcast:
			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(client.send)
			// Add queued chat messages to the current websocket message.
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			log.Println("ping....")
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Group, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	u1 := uuid.Must(uuid.NewV4())
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
