package database

import (
	"context"
	"log"
	"os"

	"rtglabs-go/ent"

	_ "github.com/mattn/go-sqlite3"
)

func NewEntClient() *ent.Client {
	dburl := os.Getenv("BLUEPRINT_DB_URL")

	client, err := ent.Open("sqlite3", dburl)
	if err != nil {
		log.Fatalf("failed opening sqlite db: %v", err)
	}

	// Run schema migration (optional, only in dev)
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	return client
}
