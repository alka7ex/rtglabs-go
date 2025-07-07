package database

import (
	"database/sql"
	"fmt"
	"time"

	// Import the SQL driver you are using
	_ "github.com/lib/pq" // For PostgreSQL
)

// Service remains conceptual; you might remove it or redefine it
// if your database interactions become direct with *sql.DB
type Service interface {
	Health() map[string]string
	// Add other methods here that your application needs,
	// which will now internally use *sql.DB and squirrel.
}

// A simple implementation of Service that wraps a *sql.DB
type sqlDBService struct {
	db *sql.DB
}

func New() Service {
	// This function might become less relevant if you're directly
	// injecting *sql.DB. Or, it could wrap the *sql.DB.
	// For now, it's a placeholder.
	// The *sql.DB will be passed directly to handlers instead of this.
	return &sqlDBService{} // You might need to pass the *sql.DB here if you keep this abstraction
}

func (s *sqlDBService) Health() map[string]string {
	// Implement health check for *sql.DB
	err := s.db.Ping() // Assuming s.db is initialized
	if err != nil {
		return map[string]string{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}
	return map[string]string{
		"status": "healthy",
	}
}

// NewSQLClient initializes and returns a new *sql.DB client.
func NewSQLClient(driverName, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Optional: Configure connection pool settings
	db.SetMaxOpenConns(25)                 // Max number of open connections
	db.SetMaxIdleConns(10)                 // Max number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Max lifetime of a connection

	return db, nil
}
