package handlers

import (
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"strconv"
	"time"
)

type BodyweightHandler struct {
	Client *ent.Client
}

func NewBodyweightHandler(client *ent.Client) *BodyweightHandler {
	return &BodyweightHandler{Client: client}
}

// --- Helper Functions ---

// toBodyweightResponse converts an ent.Bodyweight entity to a dto.BodyweightResponse DTO.
func toBodyweightResponse(bw *ent.Bodyweight) dto.BodyweightResponse {
	// Safely dereference the pointer fields, assigning a zero value if they are nil.
	// This prevents a runtime panic.
	var createdAt time.Time

	// This part is from our previous fix, converting string to int.
	unit, err := strconv.Atoi(bw.Unit)
	if err != nil {
		// Log the error if the database value is not a valid integer.
		unit = 0
	}

	// DeletedAt can be a pointer in the DTO, so we can assign it directly.
	var deletedAt *time.Time
	if bw.DeletedAt != nil {
		deletedAt = bw.DeletedAt
	}

	return dto.BodyweightResponse{
		ID:        bw.ID,
		UserID:    bw.Edges.User.ID,
		Weight:    bw.Weight,
		Unit:      unit,
		CreatedAt: createdAt,
		DeletedAt: deletedAt,
	}
}
