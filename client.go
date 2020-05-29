package main

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

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
	pingPeriod = 5 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	MSG_TEXT   = 1
	MSG_SIGNIN = 2
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
	//group id
	group_id string
	//send buffer
	send []byte
	//group
	group *Group
	//is worker
	is_worker bool
	lock      sync.Mutex
}

type Msg struct {
	UserId     string `json:"user_id"`
	MsgType    int    `json:"msg_type"`
	MsgContent string `json:"msg"`
}

type MsgResp struct {
	ErrCode string `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		log.Println("Read clean......")
		if c.is_worker {
			c.group.worker_unregister <- c
		} else {
			c.group.client_unregister <- c
		}
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.send = []byte(c.handler(message))
		c.group.read_user <- c
	}
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) handler(message []byte) string {
	var msg Msg
	if err := json.Unmarshal(message, &msg); err != nil {
		log.Println("json Unmarshal Error")
	}
	switch msg.MsgType {
	case MSG_TEXT:
	case MSG_SIGNIN:
	}
	return msg.MsgContent
}

//writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Println("Write clean......")
		if c.is_worker {
			c.group.worker_unregister <- c
		} else {
			c.group.client_unregister <- c
		}
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case client := <-c.group.write_user:
			log.Println("==========>")
			c.lock.Lock()
			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(client.send)
			w.Write(newline)
			if err := w.Close(); err != nil {
				return
			}
			c.lock.Unlock()

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
