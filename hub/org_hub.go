// Package hub provides the core WebSocket hub functionality for managing
// organizations, groups, and clients in a hierarchical messaging system.
package hub

import (
	"fmt"
	"sync"
)

// Org represents an organization that contains multiple groups.
// Each organization is a tenant in the multi-tenant system.
type Org struct {
	ID     string               `json:"id"`
	Name   string               `json:"name"`
	Groups map[string]*GroupHub `json:"groups"`
}

// OrgHub manages all organizations and their group hubs.
// It acts as the top-level hub that coordinates message routing
// across all organizations and groups in the system.
type OrgHub struct {
	Organizations     map[string]*Org    // Map of organization ID to Org
	DirectConnections map[string]*Client // Map of user ID to connected client for DMs
	Register          chan *GroupHub     // Channel for registering new groups
	Unregister        chan *GroupHub     // Channel for unregistering groups
	RegisterDM        chan *Client       // Channel for registering DM clients
	UnregisterDM      chan *Client       // Channel for unregistering DM clients
	mu                sync.RWMutex       // Mutex for thread-safe access to Organizations
	dmMu              sync.RWMutex       // Mutex for thread-safe access to DirectConnections
}

// NewOrgHub creates and initializes a new organization hub.
// It should be called once at application startup.
func NewOrgHub() *OrgHub {
	return &OrgHub{
		Organizations:     make(map[string]*Org),
		DirectConnections: make(map[string]*Client),
		Register:          make(chan *GroupHub),
		Unregister:        make(chan *GroupHub),
		RegisterDM:        make(chan *Client),
		UnregisterDM:      make(chan *Client),
	}
}

// Run handles registration and unregistration for the organization hub.
// This method should be called in a goroutine and will run continuously until the
// application shuts down.
//
// It handles four types of operations:
// 1. Register: Adds a new group to an organization (creates org if needed)
// 2. Unregister: Removes a group from an organization (removes org if empty)
// 3. RegisterDM: Registers a client for direct messaging
// 4. UnregisterDM: Unregisters a client from direct messaging
func (o *OrgHub) Run() {
	for {
		select {
		case group := <-o.Register:
			o.mu.Lock()
			org, exists := o.Organizations[group.OrgID]
			if !exists {
				org = &Org{
					ID:     group.OrgID,
					Name:   group.OrgID, // You might want to set this properly
					Groups: make(map[string]*GroupHub),
				}
				o.Organizations[group.OrgID] = org
			}
			org.Groups[group.GroupID] = group
			o.mu.Unlock()
			fmt.Printf("Group %s registered under organization: %s\n", group.GroupID, group.OrgID)

		case group := <-o.Unregister:
			o.mu.Lock()
			if org, exists := o.Organizations[group.OrgID]; exists {
				delete(org.Groups, group.GroupID)
				if len(org.Groups) == 0 {
					delete(o.Organizations, group.OrgID)
				}
			}
			o.mu.Unlock()
			fmt.Printf("Group %s unregistered from organization: %s\n", group.GroupID, group.OrgID)

		case client := <-o.RegisterDM:
			o.dmMu.Lock()
			o.DirectConnections[client.ID] = client
			o.dmMu.Unlock()
			fmt.Printf("Client %s registered for direct messaging\n", client.ID)

		case client := <-o.UnregisterDM:
			o.dmMu.Lock()
			if _, exists := o.DirectConnections[client.ID]; exists {
				delete(o.DirectConnections, client.ID)
				close(client.Send)
			}
			o.dmMu.Unlock()
			fmt.Printf("Client %s unregistered from direct messaging\n", client.ID)
		}
	}
}

// GetOrganizations returns a copy of all organizations (thread-safe).
func (o *OrgHub) GetOrganizations() map[string]*Org {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Return a copy to prevent external modification
	orgs := make(map[string]*Org, len(o.Organizations))
	for id, org := range o.Organizations {
		orgs[id] = org
	}
	return orgs
}

// GetOrganization returns a specific organization by ID (thread-safe).
func (o *OrgHub) GetOrganization(orgID string) (*Org, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()
	org, exists := o.Organizations[orgID]
	return org, exists
}

// CreateOrganization creates a new organization (thread-safe).
func (o *OrgHub) CreateOrganization(orgID, name string) *Org {
	o.mu.Lock()
	defer o.mu.Unlock()

	if org, exists := o.Organizations[orgID]; exists {
		return org
	}

	org := &Org{
		ID:     orgID,
		Name:   name,
		Groups: make(map[string]*GroupHub),
	}
	o.Organizations[orgID] = org
	return org
}

// GetGroup returns a specific group from an organization (thread-safe).
func (o *OrgHub) GetGroup(orgID, groupID string) (*GroupHub, bool) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	org, exists := o.Organizations[orgID]
	if !exists {
		return nil, false
	}

	group, exists := org.Groups[groupID]
	return group, exists
}

// BroadcastToOrg sends a message to all groups in an organization (thread-safe).
func (o *OrgHub) BroadcastToOrg(orgID string, message *Message) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if org, exists := o.Organizations[orgID]; exists {
		for _, group := range org.Groups {
			// Non-blocking send to avoid deadlock
			select {
			case group.Broadcast <- message:
			default:
				fmt.Printf("Warning: Group %s broadcast channel is full\n", group.GroupID)
			}
		}
	}
}

// BroadcastToGroup sends a message to a specific group (thread-safe).
func (o *OrgHub) BroadcastToGroup(orgID, groupID string, message *Message) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	if org, exists := o.Organizations[orgID]; exists {
		if group, exists := org.Groups[groupID]; exists {
			select {
			case group.Broadcast <- message:
			default:
				fmt.Printf("Warning: Group %s broadcast channel is full\n", groupID)
			}
		}
	}
}

// GetDirectClient returns a connected client by user ID for DM (thread-safe).
func (o *OrgHub) GetDirectClient(userID string) (*Client, bool) {
	o.dmMu.RLock()
	defer o.dmMu.RUnlock()
	client, exists := o.DirectConnections[userID]
	return client, exists
}

// SendDirectMessage sends a message directly to a specific user (thread-safe).
func (o *OrgHub) SendDirectMessage(recipientID string, message *Message) bool {
	o.dmMu.RLock()
	defer o.dmMu.RUnlock()

	if client, exists := o.DirectConnections[recipientID]; exists {
		select {
		case client.Send <- message:
			return true
		default:
			fmt.Printf("Warning: Client %s send channel is full\n", recipientID)
			return false
		}
	}
	return false
}

// GetConnectedDMUsers returns a list of all user IDs currently connected for DM (thread-safe).
func (o *OrgHub) GetConnectedDMUsers() []string {
	o.dmMu.RLock()
	defer o.dmMu.RUnlock()

	users := make([]string, 0, len(o.DirectConnections))
	for userID := range o.DirectConnections {
		users = append(users, userID)
	}
	return users
}
