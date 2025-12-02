package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-realtime-workspace/config"
	"go-realtime-workspace/database"
	"go-realtime-workspace/hub"
	"go-realtime-workspace/repository"
	"go-realtime-workspace/router"
)

func main() {
	// Load configuration
	cfg := config.DefaultConfig()

	// Initialize PostgreSQL
	pgDB, err := database.NewPostgresDB(cfg.PostgreSQL)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgDB.Close()
	log.Println("Connected to PostgreSQL")

	// Initialize Redis
	redisClient, err := database.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Connected to Redis")

	// Initialize repositories
	userRepo := repository.NewUserRepository(pgDB.DB)
	taskRepo := repository.NewTaskRepository(pgDB.DB)
	messageRepo := repository.NewMessageRepository(redisClient.Client, cfg.Redis)

	// Create the main organization hub
	orgHub := hub.NewOrgHub()
	go orgHub.Run()

	// Set up the router with all routes and middleware
	routerCfg := &router.Config{
		OrgHub:      orgHub,
		UserRepo:    userRepo,
		TaskRepo:    taskRepo,
		MessageRepo: messageRepo,
		PgHealth:    pgDB,
		RedisHealth: redisClient,
	}
	r := router.Setup(routerCfg)

	// Configure the server
	address := cfg.Server.Address
	server := &http.Server{
		Addr:         address,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start the server in a goroutine
	go func() {
		log.Printf("Server is running on %s", address)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}
