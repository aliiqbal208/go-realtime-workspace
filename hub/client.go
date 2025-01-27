package hub

import (
	"github.com/gorilla/websocket"
)

// Client represents a WebSocket client
type Client struct {
	ID    string
	Conn  *websocket.Conn
	Group *GroupHub
	Send  chan *Message
}

// writePump sends messages to the client's WebSocket connection
func (c *Client) writePump() {
	for message := range c.Send {
		c.Conn.WriteJSON(message)
	}
}

// readPump reads messages from the client's WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.Group.RemoveClient(c)
		c.Conn.Close()
	}()
	for {
		var msg Message
		if err := c.Conn.ReadJSON(&msg); err != nil {
			break
		}
		c.Group.Broadcast <- &msg
	}
}
