// internal/database/ent.go
package database

import (
	"log"
	"os"

	"rtglabs-go/ent" // Assuming this is your generated Ent client path

	_ "github.com/mattn/go-sqlite3" // Import for SQLite driver
)

func NewEntClient() *ent.Client {
	dburl := os.Getenv("BLUEPRINT_DB_URL")
	if dburl == "" {
		log.Fatalf("BLUEPRINT_DB_URL environment variable is not set for Ent client.")
	}

	client, err := ent.Open("sqlite3", dburl)
	if err != nil {
		log.Fatalf("failed opening sqlite db with URL %s: %v", dburl, err)
	}

	// REMOVE THIS LINE:
	// if err := client.Schema.Create(context.Background()); err != nil {
	// 	log.Fatalf("failed creating schema resources: %v", err)
	// }
	//
	// Explanation: Atlas will now manage your schema migrations.
	// The client.Schema.Create() method is for automatic migrations,
	// which conflicts with Atlas's versioned migration system.
	// Your application code should not attempt to manage the schema directly.

	return client
}
