// internal/database/ent.go
package database

import (
	"log"
	"os"

	"rtglabs-go/ent" // Assuming this is your generated Ent client path

	_ "github.com/lib/pq" // Import for PostgreSQL driver
)

func NewEntClient() *ent.Client {
	dburl := os.Getenv("DATABASE_URL")
	if dburl == "" {
		log.Fatalf("DATABASE_URL environment variable is not set for Ent client.")
	}

	// For PostgreSQL, the driver name is "postgres"
	client, err := ent.Open("postgres", dburl)
	if err != nil {
		log.Fatalf("failed opening PostgreSQL db with URL %s: %v", dburl, err)
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

