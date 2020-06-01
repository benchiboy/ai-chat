package main

import (
	"time"

	"github.com/gorilla/websocket"
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

const (
	//Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	//Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	//Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = 5 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512

	MSG_TEXT = 1

	MSG_USER_LIST = 5

	MSG_USER_ID = 2
)

//client node define
type Client struct {
	//client conn
	conn *websocket.Conn
	//client id
	from_user_id string
	//send buffer
	send []byte
	//group
	group *Group
	//is worker
	is_worker bool
}

type MsgHeader struct {
	MsgType int `json:"msg_type"`
}

type MsgText struct {
	FromUserId string `json:"from_user_id"`
	ToUserId   string `json:"to_user_id"`
	MsgType    int    `json:"msg_type"`
	MsgContent string `json:"msg_content"`
}

type MsgGetUsers struct {
	MsgType    int    `json:"msg_type"`
	FromUserId string `json:"from_user_id"`
	ToUserId   string `json:"to_user_id"`
	Users      []User `json:"users"`
}

type MsgGetUserId struct {
	MsgType    int    `json:"msg_type"`
	FromUserId string `json:"from_user_id"`
	ToUserId   string `json:"to_user_id"`
}

type User struct {
	UserId   string `json:"user_id"`
	UserName string `json:"user_name"`
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	PrintHead("ReadPump")
	defer func() {
		if c.is_worker {
			c.group.worker_unregister <- c
		} else {
			c.group.client_unregister <- c
		}
		PrintInfo("ReadPump Close")
		c.conn.Close()
		PrintTail("ReadPump")
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				PrintInfo(err)
			}
			break
		}
		PrintInfo("原始内容", string(message))
		c.send = message
		c.group.read_user <- c
	}
}

//writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	PrintHead("WritePump")
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		if c.is_worker {
			c.group.worker_unregister <- c
		} else {
			c.group.client_unregister <- c
		}
		ticker.Stop()
		PrintInfo("WritePump Close")
		c.conn.Close()
		PrintTail("WritePump")
	}()
	for {
		select {
		case client := <-c.group.write_user:
			lock.Lock()
			w, err := client.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				PrintInfo(err)
				return
			}
			w.Write(client.send)
			w.Write(newline)
			if err := w.Close(); err != nil {
				PrintInfo(err)
				lock.Unlock()
				return
			}
			lock.Unlock()
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
