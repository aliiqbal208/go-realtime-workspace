// Package database provides database connection and initialization.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"go-realtime-workspace/config"
	"time"

	_ "github.com/lib/pq"
)

// PostgresDB wraps the sql.DB connection.
type PostgresDB struct {
	*sql.DB
}

// NewPostgresDB creates a new PostgreSQL database connection.
func NewPostgresDB(cfg config.PostgreSQLConfig) (*PostgresDB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	return &PostgresDB{db}, nil
}

// Close closes the database connection.
func (db *PostgresDB) Close() error {
	return db.DB.Close()
}

// HealthCheck performs a database health check.
func (db *PostgresDB) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}
