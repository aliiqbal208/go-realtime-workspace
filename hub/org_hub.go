package hub

import (
	"fmt"
	"sync"
)

type Org struct {
	ID    string             `json:"id"`
	Name  string             `json:"name"`
	Group map[string]*Client `json:"groups"`
}

// OrgHub manages all organizations and their group hubs
type OrgHub struct {
	Organizations map[string]*Org
	Register      chan *GroupHub
	Unregister    chan *GroupHub
	Broadcast     chan *Message
	mu            sync.Mutex
}

// NewOrgHub creates a new organization hub
func NewOrgHub() *OrgHub {
	return &OrgHub{
		Organizations: make(map[string]*OrgHub),
		Register:      make(chan *GroupHub),
		Unregister:    make(chan *GroupHub),
		Broadcast:     make(chan *Message),
	}
}

// Run handles registering and broadcasting for the organization hub
func (o *OrgHub) Run() {
	for {
		select {
		case group := <-o.Register:
			o.mu.Lock()
			o.Organizations[group.OrgID] = group
			o.mu.Unlock()
			fmt.Printf("Group registered under organization: %s\n", group.OrgID)

		case group := <-o.Unregister:
			o.mu.Lock()
			delete(o.Organizations, group.OrgID)
			o.mu.Unlock()
			fmt.Printf("Group unregistered from organization: %s\n", group.OrgID)

		case message := <-o.Broadcast:
			o.mu.Lock()
			for _, group := range o.Organizations {
				group.Broadcast <- message
			}
			o.mu.Unlock()
		}
	}
}
