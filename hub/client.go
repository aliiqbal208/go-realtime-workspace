package hub

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// writeWait is the time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// pongWait is the time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// pingPeriod is the interval for sending pings to peer. Must be less than pongWait
	pingPeriod = (pongWait * 9) / 10

	// maxMessageSize is the maximum message size allowed from peer
	maxMessageSize = 512
)

// Client represents a WebSocket client connected to a group.
// Each client has its own goroutines for reading and writing messages.
type Client struct {
	ID    string          // Unique client identifier
	Conn  *websocket.Conn // WebSocket connection
	Group *GroupHub       // Parent group hub
	Send  chan *Message   // Buffered channel for outbound messages
}

// writePump sends messages to the client's WebSocket connection.
// It runs in its own goroutine and handles:
// - Writing messages from the Send channel
// - Sending periodic ping messages for keepalive
// - Proper cleanup on connection close
//
// A goroutine running writePump is started for each client connection.
// The application ensures that there is at most one writer to a connection
// by executing all writes from this goroutine.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("Error writing message to client %s: %v", c.ID, err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Error sending ping to client %s: %v", c.ID, err)
				return
			}
		}
	}
}

// readPump reads messages from the client's WebSocket connection.
// It runs in its own goroutine and handles:
// - Reading incoming messages from the client
// - Processing pong messages for keepalive
// - Enforcing read deadlines and message size limits
// - Proper cleanup and unregistration on disconnect
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.Group.RemoveClient(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg Message
		if err := c.Conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected close error for client %s: %v", c.ID, err)
			}
			break
		}

		// Set the client ID and group ID from the connection context
		msg.ClientID = c.ID
		msg.GroupID = c.Group.GroupID
		msg.OrgID = c.Group.OrgID

		c.Group.Broadcast <- &msg
	}
}
