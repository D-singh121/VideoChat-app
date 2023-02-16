package chat

import (
	"bytes"
	"log"
	"time"

	"github.com/fasthttp/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingWait       = (pongWait * 9) / 10
	maxMessageSize = 512
)

// --------variable for additional uses .
var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
}

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// ------readPump is a method for reading incomming messeges from websocket connection .
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error { c.Conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error :%v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.Hub.broadcast <- message
	}

}

// -------and writePump for writing  messeges to the websocket connection .
func (c *Client) writePump() {
	ticker := time.NewTicker(pingWait) // ------- this  error  occures  while writing  we can use  "pngWait" insted of this .
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)

			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.Send)
			}
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}

}

func PeerChatConn(c *websocket.Conn, hub *Hub) {
	client := &Client{Hub: hub, Conn: c, Send: make(chan []byte, 256)}
	client.Hub.register <- client

	go client.writePump()
	client.readPump()

}
