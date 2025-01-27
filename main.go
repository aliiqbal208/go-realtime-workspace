package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

// Define the necessary structs
type Group struct {
	ID        string
	Name      string
	OrgID     string
	Clients   map[string]*Client
	Broadcast chan *Message
	mu        sync.Mutex
}

type Message struct {
	OrgID    string `json:"org_id"`
	GroupID  string `json:"group_id"`
	ClientID string `json:"client_id"`
	Content  string `json:"content"`
}

type Client struct {
	ID    string
	Conn  *websocket.Conn
	Group *Group
	Send  chan *Message
}

type OrgHub struct {
	Organizations map[string]*Org
	Register      chan *GroupHub
	Unregister    chan *GroupHub
	Broadcast     chan *Message
	mu            sync.Mutex
}

type Org struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Groups map[string]*Group `json:"groups"`
}

type GroupHub struct {
	Groups     map[string]*Group
	Broadcast  chan *Message
	mu         sync.Mutex
	Register   chan *GroupHub
	Unregister chan *GroupHub
}

type WebSocketHandler struct {
	OrgHub *OrgHub
}

func (orgHub *OrgHub) Run() {
	for {
		select {
		// case org := <-orgHub.Register:
		// 	// Handle new organization registration
		// 	orgHub.mu.Lock()
		// 	orgHub.Organizations[org.ID] = org // Register Org in the map by ID
		// 	orgHub.mu.Unlock()
		// 	fmt.Printf("Registered new organization: %s\n", org.ID)

		// case org := <-orgHub.Unregister:
		// 	// Handle organization unregistration
		// 	orgHub.mu.Lock()
		// 	delete(orgHub.Organizations, org.ID) // Remove Org from the map
		// 	orgHub.mu.Unlock()
		// 	fmt.Printf("Unregistered organization: %s\n", org.ID)

		case message := <-orgHub.Broadcast:
			// Broadcast the message to all groups within all organizations
			orgHub.mu.Lock()
			for _, org := range orgHub.Organizations {
				for _, group := range org.Groups {
					for _, client := range group.Clients {
						client.Send <- message
					}
				}
			}
			orgHub.mu.Unlock()
			fmt.Println("Broadcasted message to all groups")
		}
	}
}

func NewOrgHub() *OrgHub {
	return &OrgHub{
		Organizations: make(map[string]*Org),
		Register:      make(chan *GroupHub),
		Unregister:    make(chan *GroupHub),
		Broadcast:     make(chan *Message),
	}
}

// Methods for handling WebSocket connections
func (ws *WebSocketHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["orgId"]

	var groupDetails struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	// Decode the JSON body into the orgDetails struct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&groupDetails); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the input
	if groupDetails.ID == "" || groupDetails.Name == "" {
		http.Error(w, "Missing required fields: ID and Name", http.StatusBadRequest)
		return
	}

	// Generate a unique groupID
	groupID := fmt.Sprintf("group-%d", len(ws.OrgHub.Organizations[orgID].Groups)+1)
	fmt.Println("groupID", groupID)
	// Create the new group
	group := &Group{
		ID:        groupDetails.ID,
		OrgID:     orgID,
		Clients:   make(map[string]*Client),
		Broadcast: make(chan *Message),
		Name:      groupDetails.Name,
	}

	// Retrieve the organization from the hub
	org, exists := ws.OrgHub.Organizations[orgID]
	if !exists {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	// Add the group to the organization's Groups map
	org.Groups[groupID] = group

	// Broadcast a message indicating the new group has been created
	ws.OrgHub.Broadcast <- &Message{
		OrgID:   orgID,
		GroupID: groupID,
		Content: "New group created",
	}

	// Respond with a success message
	w.Write([]byte(fmt.Sprintf("Group %s created for organization %s", groupID, orgID)))
}

func (ws *WebSocketHandler) CreateOrg(w http.ResponseWriter, r *http.Request) {
	// Parse the incoming request body to extract organization details
	var orgDetails struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	// Decode the JSON body into the orgDetails struct
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&orgDetails); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate the input
	if orgDetails.ID == "" || orgDetails.Name == "" {
		http.Error(w, "Missing required fields: ID and Name", http.StatusBadRequest)
		return
	}

	// Create a new Organization struct
	org := &Org{
		ID:     orgDetails.ID,
		Name:   orgDetails.Name,
		Groups: make(map[string]*Group), // Initialize the Groups map
	}

	// Check if the organization already exists in the OrgHub
	_, exists := ws.OrgHub.Organizations[org.ID]
	if exists {
		http.Error(w, "Organization already exists", http.StatusConflict)
		return
	}

	// Register the new organization in the OrgHub
	ws.OrgHub.mu.Lock()
	ws.OrgHub.Organizations[org.ID] = org
	ws.OrgHub.mu.Unlock()

	// Broadcast a message to indicate that a new organization has been created
	ws.OrgHub.Broadcast <- &Message{
		OrgID:   org.ID,
		Content: fmt.Sprintf("New organization %s created", org.Name),
	}

	// Respond with a success message
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("Organization %s created successfully", org.Name)))
}

type OrgRes struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (ws *WebSocketHandler) GetOrgs(w http.ResponseWriter, r *http.Request) {
	orgs := make([]OrgRes, 0)

	for _, r := range ws.OrgHub.Organizations {
		orgs = append(orgs, OrgRes{
			ID:   r.ID,
			Name: r.Name,
		})
	}

	// Set the response header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Send the list of organizations as a JSON response
	if err := json.NewEncoder(w).Encode(orgs); err != nil {
		// Handle any error that occurs while encoding the response
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

}

func (ws *WebSocketHandler) JoinGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["orgId"]
	groupID := vars["groupId"]
	clientID := "client-id" // You can get client ID from a connection context, or set it dynamically

	// Find the group and organization
	org, exists := ws.OrgHub.Organizations[orgID]
	if !exists {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	group, exists := org.Groups[groupID]
	if !exists {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	client := &Client{
		ID:    clientID,
		Group: group,
		Send:  make(chan *Message),
	}

	// Add client to the group
	group.mu.Lock()
	group.Clients[clientID] = client
	group.mu.Unlock()

	// Start client read and write pumps here (handle WebSocket connections)
	w.Write([]byte(fmt.Sprintf("Client %s joined group %s", clientID, groupID)))
}

func (ws *WebSocketHandler) BroadcastOrg(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["orgId"]
	var msg Message
	// Parse message body
	// For simplicity, assume msg is already populated in the request

	// Broadcast message to all groups in the organization
	for _, org := range ws.OrgHub.Organizations[orgID].Groups {
		for _, client := range org.Clients {
			client.Send <- &msg
		}
	}
	w.Write([]byte("Message broadcasted to organization"))
}

func (ws *WebSocketHandler) BroadcastGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orgID := vars["orgId"]
	groupID := vars["groupId"]
	var msg Message
	// Parse message body
	// For simplicity, assume msg is already populated in the request

	// Broadcast message to a specific group
	org, exists := ws.OrgHub.Organizations[orgID]
	if !exists {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	group, exists := org.Groups[groupID]
	if !exists {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	for _, client := range group.Clients {
		client.Send <- &msg
	}

	w.Write([]byte("Message broadcasted to group"))
}

func (ws *WebSocketHandler) GetOrgGroups(w http.ResponseWriter, r *http.Request) {
	// Get the orgId from the URL parameters
	vars := mux.Vars(r)
	orgID := vars["orgId"]

	// Retrieve the organization from the OrgHub using the orgId
	org, exists := ws.OrgHub.Organizations[orgID]
	if !exists {
		// If the organization doesn't exist, return an error
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	// Prepare a slice to hold the group responses
	groups := make([]struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}, 0)

	// Iterate over the groups in the organization and add them to the response slice
	for groupID, group := range org.Groups {
		groups = append(groups, struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			ID:   groupID,
			Name: group.Name, // Assuming the Group struct has a Name field
		})
	}

	for _, r := range ws.OrgHub.Organizations {
		orgs = append(orgs, OrgRes{
			ID:   r.ID,
			Name: r.Name,
		})
	}

	// Set the response header to application/json
	w.Header().Set("Content-Type", "application/json")

	// Send the groups as a JSON response
	if err := json.NewEncoder(w).Encode(groups); err != nil {
		// Handle any error that occurs while encoding the response
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func main() {
	// Create the main organization hub
	orgHub := NewOrgHub()
	go orgHub.Run()

	// Initialize the WebSocket handler
	wsHandler := &WebSocketHandler{OrgHub: orgHub}

	// Set up the router
	router := mux.NewRouter()
	router.HandleFunc("/ws/createGroup/{orgId}", wsHandler.CreateGroup).Methods("POST")
	router.HandleFunc("/ws/getOrgGroups/{orgId}", wsHandler.GetOrgGroups).Methods("GET")
	router.HandleFunc("/ws/createOrg", wsHandler.CreateOrg).Methods("POST")
	router.HandleFunc("/ws/getOrgs", wsHandler.GetOrgs).Methods("GET")
	router.HandleFunc("/ws/joinGroup/{orgId}/{groupId}", wsHandler.JoinGroup)
	router.HandleFunc("/ws/broadcastOrg/{orgId}", wsHandler.BroadcastOrg).Methods("POST")
	router.HandleFunc("/ws/broadcastGroup/{orgId}/{groupId}", wsHandler.BroadcastGroup).Methods("POST")

	// Start the server
	address := ":8080"
	fmt.Printf("Server is running on %s\n", address)
	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// package main

// import (
// 	"fmt"
// 	"log"
// 	"net/http"
// 	"org-hub-architecture/handlers"
// 	"org-hub-architecture/hub"

// 	"github.com/gorilla/mux"
// )

// func main() {
// 	// Create the main organization hub
// 	orgHub := hub.NewOrgHub()
// 	go orgHub.Run()

// 	// Initialize the WebSocket handler
// 	wsHandler := handlers.NewWebSocketHandler(orgHub)

// 	// Set up the router
// 	router := mux.NewRouter()
// 	router.HandleFunc("/ws/createGroup/{orgId}", wsHandler.CreateGroup).Methods("POST")
// 	router.HandleFunc("/ws/joinGroup/{orgId}/{groupId}", wsHandler.JoinGroup)
// 	router.HandleFunc("/ws/broadcastOrg/{orgId}", wsHandler.BroadcastOrg).Methods("POST")
// 	router.HandleFunc("/ws/broadcastGroup/{orgId}/{groupId}", wsHandler.BroadcastGroup).Methods("POST")

// 	// Start the server
// 	address := ":8080"
// 	fmt.Printf("Server is running on %s\n", address)
// 	if err := http.ListenAndServe(address, router); err != nil {
// 		log.Fatalf("Failed to start server: %v", err)
// 	}
// }
