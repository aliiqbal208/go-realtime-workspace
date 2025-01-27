package hub

import (
	"sync"
)

// GroupHub manages clients for a specific group
type Group struct {
	ID        string
	OrgID     string
	Clients   map[string]*Client
	Broadcast chan *Message
	mu        sync.Mutex
}

type GroupHub struct {
	Groups     map[string]*Group
	Broadcast  chan *Message
	mu         sync.Mutex
	Register   chan *GroupHub
	Unregister chan *GroupHub
}

// Message represents a message sent within a group or organization
type Message struct {
	OrgID    string `json:"org_id"`
	GroupID  string `json:"group_id"`
	ClientID string `json:"client_id"`
	Content  string `json:"content"`
}

// NewGroupHub creates a new group hub
func NewGroupHub(orgID, groupID string) *GroupHub {
	return &GroupHub{
		OrgID:     orgID,
		GroupID:   groupID,
		Clients:   make(map[string]*Client),
		Broadcast: make(chan *Message),
	}
}

// Run handles messages for the group
func (g *GroupHub) Run() {
	for {
		select {
		case message := <-g.Broadcast:
			g.mu.Lock()
			for _, client := range g.Clients {
				client.Send <- message
			}
			g.mu.Unlock()
		}
	}
}

// AddClient adds a new client to the group
func (g *GroupHub) AddClient(client *Client) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Clients[client.ID] = client
	go client.writePump()
	go client.readPump()
}

// RemoveClient removes a client from the group
func (g *GroupHub) RemoveClient(client *Client) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.Clients, client.ID)
	close(client.Send)
}
