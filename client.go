package main

import (
	"bytes"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
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
	pingPeriod = 5 * time.Second

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
	//group id
	group_id string
	//is_worker
	is_worker bool
	//send buffer
	send []byte
	//group
	group *Group
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		log.Println("Read Disconnect.....")
		c.group.client_unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	i := 0
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
		log.Println(string(message)+"=======>", i)
		i++
		//c.send = message
		c.group.message <- string(message)
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
		select {
		case client := <-c.group.from_user:
			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(client.send)
			// Add queued chat messages to the current websocket message.
			if err := w.Close(); err != nil {
				return
			}
		case other_client := <-c.group.to_user:
			log.Println("get to_user.....", string(other_client.send))
			w, err := other_client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(other_client.send)
			// Add queued chat messages to the current websocket message.
			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
