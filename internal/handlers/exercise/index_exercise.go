// handlers/exercise.go
package handlers

import (
	"context"
	"fmt"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/provider"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/typesense/typesense-go/v3/typesense/api"
	"github.com/typesense/typesense-go/v3/typesense/api/pointer"
)

// IndexExercise handler
func (h *ExerciseHandler) IndexExercise(c echo.Context) error {
	// ... (pagination parameters and searchName extraction - no change)
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit < 1 {
		limit = 15
	}
	if limit > 100 {
		limit = 100
	}
	offset := (page - 1) * limit

	searchName := strings.TrimSpace(c.QueryParam("q"))

	var pagination provider.PaginationResponse

	searchParams := &api.SearchCollectionParams{
		Q:       pointer.String(searchName),
		QueryBy: pointer.String("name,description"),
		Page:    pointer.Int(page),
		PerPage: pointer.Int(limit),
	}

	if searchName == "" {
		searchParams.SortBy = pointer.String("created_at:desc")
	}

	tsClient := h.TypesenseClient
	searchRes, err := tsClient.Collection("exercises").Documents().Search(context.Background(), searchParams)
	if err != nil {
		fmt.Printf("ERROR: Typesense search failed: %v\n", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search exercises")
	}

	fmt.Printf("DEBUG: Typesense search successful. Total found: %d\n", *searchRes.Found)
	fmt.Printf("DEBUG: Number of hits returned: %d\n", len(*searchRes.Hits))

	if searchRes.Hits == nil || len(*searchRes.Hits) == 0 {
		fmt.Println("DEBUG: searchRes.Hits is nil or empty, returning empty response.")
		pagination = provider.GeneratePaginationData(int(*searchRes.Found), page, limit, c.Request().URL.Path, c.QueryParams())
		zero := 0
		pagination.To = &zero
		return c.JSON(http.StatusOK, dto.ListExerciseResponse{
			Data:               []dto.ExerciseResponse{},
			PaginationResponse: pagination,
		})
	}

	var exercisesResponse []dto.ExerciseResponse
	for i, hit := range *searchRes.Hits {
		fmt.Printf("DEBUG: Processing hit %d. Raw document: %+v\n", i, *hit.Document)

		document := *hit.Document

		// --- UUID Extraction ---
		uuidStr, ok := document["uuid"].(string)
		if !ok {
			fmt.Printf("WARN: Hit %d: 'uuid' field missing or not a string. Value: %+v\n", i, document["uuid"])
			continue
		}
		uid, err := uuid.Parse(uuidStr)
		if err != nil {
			fmt.Printf("WARN: Hit %d: Invalid UUID string '%s'. Error: %v\n", i, uuidStr, err)
			continue
		}

		// --- Name Extraction ---
		name, ok := document["name"].(string)
		if !ok {
			fmt.Printf("WARN: Hit %d (UUID: %s): 'name' missing or not string. Value: %+v\n", i, uuidStr, document["name"])
			continue
		}

		// --- NEW FIELD EXTRACTION ---
		// Use empty string as default if field is missing or not a string
		description, _ := document["description"].(string)
		position, _ := document["position"].(string)
		forceType, _ := document["force_type"].(string)
		difficulty, _ := document["difficulty"].(string)
		movementType, _ := document["movement_type"].(string)
		muscleGroup, _ := document["muscle_group"].(string)
		equipment, _ := document["equipment"].(string)
		bodypart, _ := document["bodypart"].(string)

		// --- Timestamps ---
		createdAtUnix, ok := document["created_at"].(float64)
		var createdAt time.Time
		if ok {
			createdAt = time.Unix(int64(createdAtUnix), 0)
		} else {
			if val, exists := document["created_at"]; exists {
				fmt.Printf("WARN: Hit %d (UUID: %s): 'created_at' not float64. Actual type: %T, Value: %+v\n", i, uuidStr, val, val)
			} else {
				fmt.Printf("WARN: Hit %d (UUID: %s): 'created_at' missing.\n", i, uuidStr)
			}
		}

		updatedAtUnix, ok := document["updated_at"].(float64)
		var updatedAt time.Time
		if ok {
			updatedAt = time.Unix(int64(updatedAtUnix), 0)
		} else {
			if val, exists := document["updated_at"]; exists {
				fmt.Printf("WARN: Hit %d (UUID: %s): 'updated_at' not float64. Actual type: %T, Value: %+v\n", i, uuidStr, val, val)
			} else {
				fmt.Printf("WARN: Hit %d (UUID: %s): 'updated_at' missing.\n", i, uuidStr)
			}
		}

		var deletedAt *time.Time
		if deletedAtUnix, ok := document["deleted_at"].(float64); ok && deletedAtUnix > 0 {
			t := time.Unix(int64(deletedAtUnix), 0)
			deletedAt = &t
		} else {
			if val, exists := document["deleted_at"]; exists {
				fmt.Printf("DEBUG: Hit %d (UUID: %s): 'deleted_at' not float64 or 0. Actual type: %T, Value: %+v\n", i, uuidStr, val, val)
			} else {
				fmt.Printf("DEBUG: Hit %d (UUID: %s): 'deleted_at' missing or nil.\n", i, uuidStr)
			}
		}

		exercisesResponse = append(exercisesResponse, dto.ExerciseResponse{
			ID:           uid,
			Name:         name,
			Description:  description,  // <--- ASSIGN THIS
			Position:     position,     // <--- ASSIGN THIS
			ForceType:    forceType,    // <--- ASSIGN THIS
			Difficulty:   difficulty,   // <--- ASSIGN THIS
			MovementType: movementType, // <--- ASSIGN THIS
			MuscleGroup:  muscleGroup,  // <--- ASSIGN THIS
			Equipment:    equipment,    // <--- ASSIGN THIS
			Bodypart:     bodypart,     // <--- ASSIGN THIS
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
			DeletedAt:    deletedAt,
		})
		fmt.Printf("DEBUG: Successfully added doc %d (UUID: %s) to response slice.\n", i, uuidStr)
	}

	totalCount := 0
	if searchRes.Found != nil {
		totalCount = int(*searchRes.Found)
	}
	pagination = provider.GeneratePaginationData(totalCount, page, limit, c.Request().URL.Path, c.QueryParams())

	if len(exercisesResponse) > 0 {
		tempTo := offset + len(exercisesResponse)
		pagination.To = &tempTo
	} else {
		zero := 0
		pagination.To = &zero
	}

	return c.JSON(http.StatusOK, dto.ListExerciseResponse{
		Data:               exercisesResponse,
		PaginationResponse: pagination,
	})
}
