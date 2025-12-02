// Package handlers provides HTTP and WebSocket handlers for the organization hub.
// It exposes RESTful endpoints for organization/group management and WebSocket
// endpoints for real-time communication.
package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"go-realtime-workspace/hub"
	"go-realtime-workspace/models"
	"go-realtime-workspace/repository"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// WebSocketHandler handles WebSocket connections and HTTP requests.
// It acts as the interface between HTTP requests and the hub system.
type WebSocketHandler struct {
	OrgHub   *hub.OrgHub
	MsgRepo  *repository.MessageRepository
	UserRepo *repository.UserRepository
}

// NewWebSocketHandler creates a new WebSocket handler.
// It should be initialized with an active OrgHub instance.
func NewWebSocketHandler(orgHub *hub.OrgHub, msgRepo *repository.MessageRepository, userRepo *repository.UserRepository) *WebSocketHandler {
	return &WebSocketHandler{
		OrgHub:   orgHub,
		MsgRepo:  msgRepo,
		UserRepo: userRepo,
	}
}

// upgrader configures the WebSocket upgrader with buffer sizes and CORS settings.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // Allow all origins (configure for production)
}

// CreateOrg creates a new organization
func (h *WebSocketHandler) CreateOrg(w http.ResponseWriter, r *http.Request) {
	var orgDetails struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&orgDetails); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if orgDetails.ID == "" || orgDetails.Name == "" {
		http.Error(w, "Missing required fields: ID and Name", http.StatusBadRequest)
		return
	}

	// Check if organization already exists
	if _, exists := h.OrgHub.GetOrganization(orgDetails.ID); exists {
		http.Error(w, "Organization already exists", http.StatusConflict)
		return
	}

	// Create the organization
	org := h.OrgHub.CreateOrganization(orgDetails.ID, orgDetails.Name)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Organization %s created successfully", org.Name),
		"org_id":  org.ID,
	})
}

// GetOrgs retrieves all organizations
func (h *WebSocketHandler) GetOrgs(w http.ResponseWriter, r *http.Request) {
	type OrgResponse struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	organizations := h.OrgHub.GetOrganizations()
	orgs := make([]OrgResponse, 0, len(organizations))
	for _, org := range organizations {
		orgs = append(orgs, OrgResponse{
			ID:   org.ID,
			Name: org.Name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(orgs)
}

// CreateGroup creates a new group in an organization
func (h *WebSocketHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]

	var groupDetails struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&groupDetails); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if groupDetails.ID == "" || groupDetails.Name == "" {
		http.Error(w, "Missing required fields: ID and Name", http.StatusBadRequest)
		return
	}

	// Check if organization exists
	org, exists := h.OrgHub.GetOrganization(orgID)
	if !exists {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	// Check if group already exists
	if _, exists := org.Groups[groupDetails.ID]; exists {
		http.Error(w, "Group already exists", http.StatusConflict)
		return
	}

	// Create and start the group hub
	group := hub.NewGroupHub(orgID, groupDetails.ID)
	group.Name = groupDetails.Name

	// Add group to organization (we need a method for this)
	h.OrgHub.Register <- group

	// Start the group hub
	go group.Run()

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status":   "success",
		"message":  fmt.Sprintf("Group %s created successfully", groupDetails.Name),
		"group_id": groupDetails.ID,
		"org_id":   orgID,
	})
}

// GetOrgGroups retrieves all groups in an organization
func (h *WebSocketHandler) GetOrgGroups(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]

	type GroupResponse struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	org, exists := h.OrgHub.GetOrganization(orgID)
	if !exists {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	groups := make([]GroupResponse, 0, len(org.Groups))
	for _, group := range org.Groups {
		groups = append(groups, GroupResponse{
			ID:   group.GroupID,
			Name: group.Name,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(groups)
}

// JoinGroup adds a client to a specific group via WebSocket
func (h *WebSocketHandler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	groupID := mux.Vars(r)["groupId"]
	clientID := r.URL.Query().Get("clientId")

	if clientID == "" {
		http.Error(w, "clientId query parameter is required", http.StatusBadRequest)
		return
	}

	// Check if group exists
	group, exists := h.OrgHub.GetGroup(orgID, groupID)
	if !exists {
		http.Error(w, "Organization or group not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v", err)
		return
	}

	client := &hub.Client{
		ID:    clientID,
		Conn:  conn,
		Group: group,
		Send:  make(chan *hub.Message, 256),
	}

	group.AddClient(client)
	log.Printf("Client %s joined group %s in organization %s", clientID, groupID, orgID)
}

// BroadcastOrg sends a message to all groups in the specified organization
func (h *WebSocketHandler) BroadcastOrg(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]

	// Check if organization exists
	_, exists := h.OrgHub.GetOrganization(orgID)
	if !exists {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	var message hub.Message
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		http.Error(w, "Invalid message format", http.StatusBadRequest)
		return
	}

	message.OrgID = orgID

	// Use the OrgHub broadcast method
	h.OrgHub.BroadcastToOrg(orgID, &message)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Message broadcasted to organization"})
}

// BroadcastGroup sends a message to all clients in a specific group within an organization
func (h *WebSocketHandler) BroadcastGroup(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	groupID := mux.Vars(r)["groupId"]

	// Check if group exists
	_, exists := h.OrgHub.GetGroup(orgID, groupID)
	if !exists {
		http.Error(w, "Organization or group not found", http.StatusNotFound)
		return
	}

	var message hub.Message
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		http.Error(w, "Invalid message format", http.StatusBadRequest)
		return
	}

	message.OrgID = orgID
	message.GroupID = groupID

	// Persist message to Redis
	if h.MsgRepo != nil {
		chatMsg := models.ChatMessage{
			OrgID:     message.OrgID,
			GroupID:   message.GroupID,
			ClientID:  message.ClientID,
			Content:   message.Content,
			Timestamp: time.Now(),
		}

		// Get username if UserRepo is available
		if h.UserRepo != nil && message.ClientID != "" {
			if user, err := h.UserRepo.GetByID(context.Background(), message.ClientID); err == nil {
				chatMsg.Username = user.Username
			}
		}

		if err := h.MsgRepo.Save(context.Background(), chatMsg); err != nil {
			log.Printf("Error saving message to Redis: %v", err)
			// Don't fail the request if Redis save fails
		}
	}

	// Use the OrgHub broadcast method
	h.OrgHub.BroadcastToGroup(orgID, groupID, &message)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Message broadcasted to group"})
}

// ConnectDM establishes a WebSocket connection for direct messaging
func (h *WebSocketHandler) ConnectDM(w http.ResponseWriter, r *http.Request) {
	userID := mux.Vars(r)["userId"]

	if userID == "" {
		http.Error(w, "userId is required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v", err)
		return
	}

	// Create a client for DM (Group is nil for DM clients)
	client := &hub.Client{
		ID:    userID,
		Conn:  conn,
		Group: nil, // DM clients don't belong to a group
		Send:  make(chan *hub.Message, 256),
	}

	// Register with OrgHub for DM
	h.OrgHub.RegisterDM <- client

	// Start read and write pumps
	go client.WritePump()
	go h.readPumpDM(client)

	log.Printf("Client %s connected for direct messaging", userID)
}

// readPumpDM handles incoming DM messages from WebSocket
func (h *WebSocketHandler) readPumpDM(client *hub.Client) {
	defer func() {
		h.OrgHub.UnregisterDM <- client
		client.Conn.Close()
	}()

	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		var message hub.Message
		err := client.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Set sender ID and timestamp
		message.ClientID = client.ID
		message.Timestamp = time.Now()

		// Persist DM to Redis
		if h.MsgRepo != nil && message.RecipientID != "" {
			chatMsg := models.ChatMessage{
				OrgID:       "dm", // Special org ID for direct messages
				GroupID:     h.getDMRoomID(client.ID, message.RecipientID),
				ClientID:    message.ClientID,
				Content:     message.Content,
				Timestamp:   message.Timestamp,
				RecipientID: message.RecipientID,
			}

			// Get username if available
			if h.UserRepo != nil {
				if user, err := h.UserRepo.GetByID(context.Background(), client.ID); err == nil {
					chatMsg.Username = user.Username
				}
			}

			if err := h.MsgRepo.Save(context.Background(), chatMsg); err != nil {
				log.Printf("Error saving DM to Redis: %v", err)
			}
		}

		// Send message to recipient
		if message.RecipientID != "" {
			sent := h.OrgHub.SendDirectMessage(message.RecipientID, &message)
			if !sent {
				log.Printf("Failed to send DM to %s (user not connected)", message.RecipientID)
			}
		}
	}
}

// SendDM sends a direct message to another user via REST API
func (h *WebSocketHandler) SendDM(w http.ResponseWriter, r *http.Request) {
	senderID := mux.Vars(r)["userId"]
	recipientID := mux.Vars(r)["recipientId"]

	var message hub.Message
	if err := json.NewDecoder(r.Body).Decode(&message); err != nil {
		http.Error(w, "Invalid message format", http.StatusBadRequest)
		return
	}

	message.ClientID = senderID
	message.RecipientID = recipientID
	message.Timestamp = time.Now()

	// Persist DM to Redis
	if h.MsgRepo != nil {
		chatMsg := models.ChatMessage{
			OrgID:       "dm",
			GroupID:     h.getDMRoomID(senderID, recipientID),
			ClientID:    senderID,
			Content:     message.Content,
			Timestamp:   message.Timestamp,
			RecipientID: recipientID,
		}

		// Get username if available
		if h.UserRepo != nil {
			if user, err := h.UserRepo.GetByID(context.Background(), senderID); err == nil {
				chatMsg.Username = user.Username
			}
		}

		if err := h.MsgRepo.Save(context.Background(), chatMsg); err != nil {
			log.Printf("Error saving DM to Redis: %v", err)
		}
	}

	// Send message to recipient
	sent := h.OrgHub.SendDirectMessage(recipientID, &message)
	if !sent {
		http.Error(w, "Recipient not connected", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Direct message sent"})
}

// GetDMHistory retrieves direct message history between two users
func (h *WebSocketHandler) GetDMHistory(w http.ResponseWriter, r *http.Request) {
	user1 := mux.Vars(r)["userId"]
	user2 := mux.Vars(r)["recipientId"]

	dmRoomID := h.getDMRoomID(user1, user2)

	messages, err := h.MsgRepo.GetHistory(context.Background(), "dm", dmRoomID, 100)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to retrieve DM history: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(messages)
}

// GetConnectedUsers returns a list of users currently connected for DM
func (h *WebSocketHandler) GetConnectedUsers(w http.ResponseWriter, r *http.Request) {
	users := h.OrgHub.GetConnectedDMUsers()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"connected_users": users,
		"count":           len(users),
	})
}

// getDMRoomID generates a consistent room ID for DM between two users
// Always orders user IDs alphabetically to ensure same room ID regardless of who initiates
func (h *WebSocketHandler) getDMRoomID(user1, user2 string) string {
	if user1 < user2 {
		return fmt.Sprintf("%s_%s", user1, user2)
	}
	return fmt.Sprintf("%s_%s", user2, user1)
}
