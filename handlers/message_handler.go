package handlers

import (
	"encoding/json"
	"net/http"
	"go-realtime-workspace/repository"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// MessageHandler handles message history HTTP requests.
type MessageHandler struct {
	repo *repository.MessageRepository
}

// NewMessageHandler creates a new message handler.
func NewMessageHandler(repo *repository.MessageRepository) *MessageHandler {
	return &MessageHandler{repo: repo}
}

// GetHistory retrieves message history for a group.
func (h *MessageHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	groupID := mux.Vars(r)["groupId"]

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := int64(50) // default
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			limit = l
		}
	}

	messages, err := h.repo.GetHistory(r.Context(), orgID, groupID, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}

// GetHistoryAfter retrieves messages after a specific timestamp.
func (h *MessageHandler) GetHistoryAfter(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	groupID := mux.Vars(r)["groupId"]

	// Parse after timestamp parameter
	afterStr := r.URL.Query().Get("after")
	if afterStr == "" {
		http.Error(w, "after query parameter is required (Unix timestamp)", http.StatusBadRequest)
		return
	}

	afterUnix, err := strconv.ParseInt(afterStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid after timestamp", http.StatusBadRequest)
		return
	}
	after := time.Unix(afterUnix, 0)

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := int64(50)
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			limit = l
		}
	}

	messages, err := h.repo.GetHistoryAfter(r.Context(), orgID, groupID, after, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}

// GetHistoryBetween retrieves messages between two timestamps.
func (h *MessageHandler) GetHistoryBetween(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	groupID := mux.Vars(r)["groupId"]

	// Parse start timestamp
	startStr := r.URL.Query().Get("start")
	if startStr == "" {
		http.Error(w, "start query parameter is required (Unix timestamp)", http.StatusBadRequest)
		return
	}
	startUnix, err := strconv.ParseInt(startStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid start timestamp", http.StatusBadRequest)
		return
	}

	// Parse end timestamp
	endStr := r.URL.Query().Get("end")
	if endStr == "" {
		http.Error(w, "end query parameter is required (Unix timestamp)", http.StatusBadRequest)
		return
	}
	endUnix, err := strconv.ParseInt(endStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid end timestamp", http.StatusBadRequest)
		return
	}

	start := time.Unix(startUnix, 0)
	end := time.Unix(endUnix, 0)

	// Parse limit parameter
	limitStr := r.URL.Query().Get("limit")
	limit := int64(50)
	if limitStr != "" {
		if l, err := strconv.ParseInt(limitStr, 10, 64); err == nil {
			limit = l
		}
	}

	messages, err := h.repo.GetHistoryBetween(r.Context(), orgID, groupID, start, end, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages": messages,
		"count":    len(messages),
	})
}

// GetCount retrieves the message count for a group.
func (h *MessageHandler) GetCount(w http.ResponseWriter, r *http.Request) {
	orgID := mux.Vars(r)["orgId"]
	groupID := mux.Vars(r)["groupId"]

	count, err := h.repo.Count(r.Context(), orgID, groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"count": count,
	})
}
