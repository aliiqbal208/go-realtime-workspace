// Package config provides configuration management for the application.
// It centralizes all configuration settings for the server and WebSocket connections.
package config

import (
	"time"
)

// Config holds all configuration for the application.
// Use DefaultConfig() to get a configuration with sensible defaults.
type Config struct {
	Server     ServerConfig
	WebSocket  WebSocketConfig
	PostgreSQL PostgreSQLConfig
	Redis      RedisConfig
}

// ServerConfig holds server-related configuration.
type ServerConfig struct {
	Address      string        // Server listen address (e.g., ":8080")
	ReadTimeout  time.Duration // Maximum duration for reading the entire request
	WriteTimeout time.Duration // Maximum duration before timing out writes of the response
	IdleTimeout  time.Duration // Maximum time to wait for the next request when keep-alives are enabled
}

// WebSocketConfig holds WebSocket-related configuration.
type WebSocketConfig struct {
	ReadBufferSize  int           // Size of the read buffer in bytes
	WriteBufferSize int           // Size of the write buffer in bytes
	WriteWait       time.Duration // Time allowed to write a message to the peer
	PongWait        time.Duration // Time allowed to read the next pong message from the peer
	PingPeriod      time.Duration // Send pings to peer with this period (must be less than PongWait)
	MaxMessageSize  int64         // Maximum message size allowed from peer
	MessageBuffer   int           // Size of the buffered channel for messages
}

// PostgreSQLConfig holds PostgreSQL database configuration.
type PostgreSQLConfig struct {
	Host         string        // Database host
	Port         int           // Database port
	User         string        // Database user
	Password     string        // Database password
	Database     string        // Database name
	SSLMode      string        // SSL mode (disable, require, verify-ca, verify-full)
	MaxOpenConns int           // Maximum number of open connections
	MaxIdleConns int           // Maximum number of idle connections
	MaxLifetime  time.Duration // Maximum lifetime of a connection
}

// RedisConfig holds Redis configuration.
type RedisConfig struct {
	Host        string        // Redis host
	Port        int           // Redis port
	Password    string        // Redis password (empty if no password)
	DB          int           // Redis database number
	MaxRetries  int           // Maximum number of retries
	PoolSize    int           // Maximum number of connections
	MessageTTL  time.Duration // Time-to-live for chat messages
	MaxMessages int64         // Maximum messages to store per group
}

// DefaultConfig returns the default configuration with production-ready settings.
// These values can be overridden for specific deployment environments.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Address:      ":8080",
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		WebSocket: WebSocketConfig{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			WriteWait:       10 * time.Second,
			PongWait:        60 * time.Second,
			PingPeriod:      54 * time.Second, // Must be less than PongWait
			MaxMessageSize:  512,
			MessageBuffer:   256,
		},
		PostgreSQL: PostgreSQLConfig{
			Host:         "localhost",
			Port:         5432,
			User:         "postgres",
			Password:     "postgres",
			Database:     "realtime_workspace",
			SSLMode:      "disable",
			MaxOpenConns: 25,
			MaxIdleConns: 5,
			MaxLifetime:  5 * time.Minute,
		},
		Redis: RedisConfig{
			Host:        "localhost",
			Port:        6379,
			Password:    "",
			DB:          0,
			MaxRetries:  3,
			PoolSize:    10,
			MessageTTL:  7 * 24 * time.Hour, // 7 days
			MaxMessages: 1000,               // Keep last 1000 messages per group
		},
	}
}
