package handlers

import (
	"database/sql" // Import for sql.DB, sql.Null* types, sql.ErrNoRows
	"fmt"
	"strconv"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/model" // Import your model package
	// REMOVE: "rtglabs-go/ent" // Remove Ent import
	"github.com/Masterminds/squirrel" // Import squirrel
)

type BodyweightHandler struct {
	DB *sql.DB                       // Use standard SQL DB
	sq squirrel.StatementBuilderType // Use squirrel builder
}

func NewBodyweightHandler(db *sql.DB) *BodyweightHandler { // Accept *sql.DB
	// Initialize squirrel with the appropriate placeholder format for PostgreSQL
	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	return &BodyweightHandler{
		DB: db,
		sq: sq,
	}
}

// --- Helper Functions ---

// toBodyweightResponse converts a model.Bodyweight entity to a dto.BodyweightResponse DTO.
// NOTE: This helper assumes model.Bodyweight already contains the UserID directly,
// and doesn't rely on Ent's Edges.User.ID.
func toBodyweightResponse(bw *model.Bodyweight) dto.BodyweightResponse {
	// Safely dereference the pointer fields, assigning a zero value if they are nil.
	var deletedAt *time.Time
	if bw.DeletedAt != nil {
		deletedAt = bw.DeletedAt
	}

	// Assuming Unit in model.Bodyweight is string based on your DDL,
	// and needs to be converted to int for DTO.
	// If model.Bodyweight.Unit is already int, remove strconv.Atoi
	unit, err := strconv.Atoi(bw.Unit)
	if err != nil {
		// Log the error if the database value is not a valid integer.
		// Or handle it based on your application's error strategy.
		// For now, default to 0.
		// You might want to remove this conversion if your DB 'unit' column is INTEGER.
		fmt.Printf("Warning: Could not convert Bodyweight.Unit '%s' to int: %v\n", bw.Unit, err)
		unit = 0
	}

	return dto.BodyweightResponse{
		ID:        bw.ID,
		UserID:    bw.UserID, // Direct access to UserID from model.Bodyweight
		Weight:    bw.Weight,
		Unit:      unit,
		CreatedAt: bw.CreatedAt, // Assuming CreatedAt is already time.Time in model
		UpdatedAt: bw.UpdatedAt, // Assuming UpdatedAt is already time.Time in model
		DeletedAt: deletedAt,
	}
}
