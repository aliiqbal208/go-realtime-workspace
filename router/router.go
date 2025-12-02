// Package router provides HTTP routing configuration and middleware setup.
package router

import (
	"context"
	"net/http"
	"go-realtime-workspace/handlers"
	"go-realtime-workspace/hub"
	"go-realtime-workspace/repository"

	"github.com/gorilla/mux"
)

// Config holds the dependencies needed for router setup.
type Config struct {
	OrgHub      *hub.OrgHub
	UserRepo    *repository.UserRepository
	TaskRepo    *repository.TaskRepository
	MessageRepo *repository.MessageRepository
	PgHealth    PgHealthChecker
	RedisHealth RedisHealthChecker
}

// PgHealthChecker defines the interface for PostgreSQL health checking.
type PgHealthChecker interface {
	HealthCheck() error
}

// RedisHealthChecker defines the interface for Redis health checking.
type RedisHealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// Setup configures all routes and returns a configured router.
func Setup(cfg *Config) *mux.Router {
	router := mux.NewRouter()

	// Initialize handlers
	wsHandler := handlers.NewWebSocketHandler(cfg.OrgHub, cfg.MessageRepo, cfg.UserRepo)
	userHandler := handlers.NewUserHandler(cfg.UserRepo)
	taskHandler := handlers.NewTaskHandler(cfg.TaskRepo)
	messageHandler := handlers.NewMessageHandler(cfg.MessageRepo)

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Health check endpoint
	api.HandleFunc("/health", healthCheckHandler(cfg.PgHealth, cfg.RedisHealth)).Methods("GET")

	// Organization routes
	api.HandleFunc("/orgs", wsHandler.CreateOrg).Methods("POST")
	api.HandleFunc("/orgs", wsHandler.GetOrgs).Methods("GET")
	api.HandleFunc("/orgs/{orgId}/groups", wsHandler.CreateGroup).Methods("POST")
	api.HandleFunc("/orgs/{orgId}/groups", wsHandler.GetOrgGroups).Methods("GET")

	// Broadcast routes
	api.HandleFunc("/orgs/{orgId}/broadcast", wsHandler.BroadcastOrg).Methods("POST")
	api.HandleFunc("/orgs/{orgId}/groups/{groupId}/broadcast", wsHandler.BroadcastGroup).Methods("POST")

	// Message history routes
	api.HandleFunc("/orgs/{orgId}/groups/{groupId}/messages", messageHandler.GetHistory).Methods("GET")
	api.HandleFunc("/orgs/{orgId}/groups/{groupId}/messages/after", messageHandler.GetHistoryAfter).Methods("GET")
	api.HandleFunc("/orgs/{orgId}/groups/{groupId}/messages/between", messageHandler.GetHistoryBetween).Methods("GET")
	api.HandleFunc("/orgs/{orgId}/groups/{groupId}/messages/count", messageHandler.GetCount).Methods("GET")

	// User routes
	api.HandleFunc("/users", userHandler.Create).Methods("POST")
	api.HandleFunc("/users/{id}", userHandler.GetByID).Methods("GET")
	api.HandleFunc("/users/{id}", userHandler.Update).Methods("PUT")
	api.HandleFunc("/users/{id}", userHandler.Delete).Methods("DELETE")
	api.HandleFunc("/users/search", userHandler.GetByUsername).Methods("GET")
	api.HandleFunc("/orgs/{orgId}/users", userHandler.GetByOrg).Methods("GET")

	// Task routes
	api.HandleFunc("/users/{userId}/tasks", taskHandler.Create).Methods("POST")
	api.HandleFunc("/users/{userId}/tasks", taskHandler.GetByUser).Methods("GET")
	api.HandleFunc("/users/{userId}/tasks/due-soon", taskHandler.GetDueSoon).Methods("GET")
	api.HandleFunc("/tasks/{id}", taskHandler.GetByID).Methods("GET")
	api.HandleFunc("/tasks/{id}", taskHandler.Update).Methods("PUT")
	api.HandleFunc("/tasks/{id}", taskHandler.Delete).Methods("DELETE")

	// Direct Messaging routes
	api.HandleFunc("/dm/{userId}/{recipientId}", wsHandler.SendDM).Methods("POST")
	api.HandleFunc("/dm/{userId}/{recipientId}/history", wsHandler.GetDMHistory).Methods("GET")
	api.HandleFunc("/dm/connected-users", wsHandler.GetConnectedUsers).Methods("GET")

	// WebSocket routes
	router.HandleFunc("/ws/orgs/{orgId}/groups/{groupId}", wsHandler.JoinGroup)
	router.HandleFunc("/ws/dm/{userId}", wsHandler.ConnectDM)

	return router
}

// healthCheckHandler creates a handler for health check endpoints.
func healthCheckHandler(pgHealth PgHealthChecker, redisHealth RedisHealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check PostgreSQL
		if err := pgHealth.HealthCheck(); err != nil {
			http.Error(w, "PostgreSQL unhealthy", http.StatusServiceUnavailable)
			return
		}

		// Check Redis
		if err := redisHealth.HealthCheck(r.Context()); err != nil {
			http.Error(w, "Redis unhealthy", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
