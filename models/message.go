package models

import (
	"time"
)

// ChatMessage represents a stored chat message in Redis.
type ChatMessage struct {
	ID          string    `json:"id"`
	OrgID       string    `json:"org_id"`
	GroupID     string    `json:"group_id"`
	ClientID    string    `json:"client_id"`
	RecipientID string    `json:"recipient_id,omitempty"` // For direct messages
	Username    string    `json:"username,omitempty"`
	Content     string    `json:"content"`
	Timestamp   time.Time `json:"timestamp"`
}
