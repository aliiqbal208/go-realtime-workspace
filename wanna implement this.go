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

type Message struct {
	OrgID    string `json:"org_id"`
	GroupID  string `json:"group_id"`
	ClientID string `json:"client_id"`
	Content  string `json:"content"`
}

type Org struct {
	ID    string             `json:"id"`
	Name  string             `json:"name"`
	Group map[string]*Client `json:"groups"`
}

type OrgHub struct {
	Organizations map[string]*Org
	Register      chan *GroupHub
	Unregister    chan *GroupHub
	Broadcast     chan *Message
	mu            sync.Mutex
}

type Client struct {
	ID    string
	Conn  *websocket.Conn
	Group *GroupHub
	Send  chan *Message
}