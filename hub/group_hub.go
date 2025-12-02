package hub

import (
	"fmt"
	"sync"
	"time"
)

// Message represents a message sent within a group or organization.
// It contains routing information and the actual message content.
type Message struct {
	OrgID       string    `json:"org_id"`       // Organization ID for routing
	GroupID     string    `json:"group_id"`     // Group ID for routing
	ClientID    string    `json:"client_id"`    // Originating client ID
	RecipientID string    `json:"recipient_id"` // Recipient ID for direct messages
	Content     string    `json:"content"`      // Message payload
	Timestamp   time.Time `json:"timestamp"`    // Message timestamp
}

// GroupHub manages clients for a specific group within an organization.
// It handles client registration, message broadcasting, and cleanup.
type GroupHub struct {
	OrgID      string             // Parent organization ID
	GroupID    string             // Unique group identifier
	Name       string             // Human-readable group name
	Clients    map[string]*Client // Map of client ID to Client
	Broadcast  chan *Message      // Channel for broadcasting messages
	Register   chan *Client       // Channel for registering clients
	Unregister chan *Client       // Channel for unregistering clients
	mu         sync.RWMutex       // Mutex for thread-safe access to Clients
}

// NewGroupHub creates and initializes a new group hub.
// The group hub must be started by calling Run() in a goroutine.
func NewGroupHub(orgID, groupID string) *GroupHub {
	return &GroupHub{
		OrgID:      orgID,
		GroupID:    groupID,
		Clients:    make(map[string]*Client),
		Broadcast:  make(chan *Message, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

// Run handles messages, client registration, and unregistration for the group.
// This method should be called in a goroutine and will run continuously.
//
// It handles three types of operations:
// 1. Register: Adds a new client to the group
// 2. Unregister: Removes a client from the group and closes their channel
// 3. Broadcast: Sends a message to all clients in the group (non-blocking)
func (g *GroupHub) Run() {
	for {
		select {
		case client := <-g.Register:
			g.mu.Lock()
			g.Clients[client.ID] = client
			g.mu.Unlock()
			fmt.Printf("Client %s joined group %s in org %s\n", client.ID, g.GroupID, g.OrgID)

		case client := <-g.Unregister:
			g.mu.Lock()
			if _, exists := g.Clients[client.ID]; exists {
				delete(g.Clients, client.ID)
				close(client.Send)
				fmt.Printf("Client %s left group %s in org %s\n", client.ID, g.GroupID, g.OrgID)
			}
			g.mu.Unlock()

		case message := <-g.Broadcast:
			g.mu.RLock()
			for _, client := range g.Clients {
				// Non-blocking send to avoid deadlock
				select {
				case client.Send <- message:
				default:
					fmt.Printf("Warning: Client %s send channel is full\n", client.ID)
				}
			}
			g.mu.RUnlock()
		}
	}
}

// AddClient adds a new client to the group and starts their read/write pumps.
// This is a convenience method that handles all the setup for a new client.
func (g *GroupHub) AddClient(client *Client) {
	g.Register <- client
	go client.WritePump()
	go client.readPump()
}

// RemoveClient removes a client from the group.
// This will trigger cleanup and close the client's send channel.
func (g *GroupHub) RemoveClient(client *Client) {
	g.Unregister <- client
}
