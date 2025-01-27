package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"org-hub-architecture/hub"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type WebSocketHandler struct {
	OrgHub *hub.OrgHub
}

func NewWebSocketHandler(orgHub *hub.OrgHub) *WebSocketHandler {
	return &WebSocketHandler{OrgHub: orgHub}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// CreateGroup creates a new group in an organization
func (h *WebSocketHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	groupID := r.URL.Query().Get("groupId")

	group := hub.NewGroupHub(orgID, groupID)
	h.OrgHub.Register <- group
	go group.Run()

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(group)
}

// JoinGroup adds a client to a specific group
func (h *WebSocketHandler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	groupID := mux.Vars(r)["groupId"]
	fmt.Println("JoinGroup groupID", groupID)
	group, exists := h.OrgHub.Organizations[orgID]
	if !exists {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade WebSocket connection", http.StatusBadRequest)
		return
	}

	client := &hub.Client{
		ID:    r.URL.Query().Get("clientId"),
		Conn:  conn,
		Group: group,
		Send:  make(chan *hub.Message),
	}

	group.AddClient(client)
}

// BroadcastOrg sends a message to all groups in the specified organization
func (h *WebSocketHandler) BroadcastOrg(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]

	// Retrieve the organization hub
	orgHub, exists := h.OrgHub.Organizations[orgID]
	if !exists {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	// Parse the incoming message from the request body
	var message hub.Message
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		http.Error(w, "Invalid message format", http.StatusBadRequest)
		return
	}

	// Add the organization ID to the message for context
	message.OrgID = orgID

	// Broadcast the message to all groups in the organization
	orgHub.Broadcast <- &message

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Message broadcasted to organization"})
}

// BroadcastGroup sends a message to all clients in a specific group within an organization
func (h *WebSocketHandler) BroadcastGroup(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	groupID := mux.Vars(r)["groupId"]

	// Retrieve the organization hub
	orgHub, exists := h.OrgHub.Organizations[orgID]
	if !exists {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	// Retrieve the specific group hub within the organization
	group, exists := orgHub.Groups[groupID]
	if !exists {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Parse the incoming message from the request body
	var message hub.Message
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		http.Error(w, "Invalid message format", http.StatusBadRequest)
		return
	}

	// Add the organization and group IDs to the message for context
	message.OrgID = orgID
	message.GroupID = groupID

	// Broadcast the message to all clients in the group
	group.Broadcast <- &message

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Message broadcasted to group"})
}
