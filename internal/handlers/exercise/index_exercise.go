// handlers/exercise.go
package handlers

import (
	"context"
	"fmt" // Keep fmt for debugging output
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/provider"
	"strconv" // <--- Ensure strconv is imported for Atoi
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/typesense/typesense-go/v3/typesense/api"
	"github.com/typesense/typesense-go/v3/typesense/api/pointer"
)

// IndexExercise handler
func (h *ExerciseHandler) IndexExercise(c echo.Context) error {
	// --- Pagination Parameters ---
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
	// --- End Pagination Parameters ---

	searchName := strings.TrimSpace(c.QueryParam("name"))

	var pagination provider.PaginationResponse

	// --- Always use Typesense for search or general listing ---
	searchParams := &api.SearchCollectionParams{
		Q:       pointer.String(searchName),
		QueryBy: pointer.String("name,description"), // Query both name and description
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

		// --- ID Extraction (MODIFIED AGAIN: Convert string to int) ---
		idStr, ok := document["id"].(string) // Still get it as string from Typesense JSON
		if !ok {
			fmt.Printf("WARN: Hit %d: 'id' missing or not string. Value: %+v\n", i, document["id"])
			continue // Skip this malformed document
		}
		// Convert the string ID to an int
		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			fmt.Printf("WARN: Hit %d (ID string: %s): failed to convert 'id' to int. Error: %v\n", i, idStr, err)
			continue // Skip this document if ID cannot be converted
		}

		// --- Name Extraction ---
		name, ok := document["name"].(string)
		if !ok {
			fmt.Printf("WARN: Hit %d (ID: %d): 'name' missing or not string. Value: %+v\n", i, idInt, document["name"])
			continue
		}

		// --- Timestamps (no changes needed here as they were already float64 assertion) ---
		createdAtUnix, ok := document["created_at"].(float64)
		var createdAt time.Time
		if ok {
			createdAt = time.Unix(int64(createdAtUnix), 0)
		} else {
			if val, exists := document["created_at"]; exists {
				fmt.Printf("WARN: Hit %d (ID: %d): 'created_at' not float64. Actual type: %T, Value: %+v\n", i, idInt, val, val)
			} else {
				fmt.Printf("WARN: Hit %d (ID: %d): 'created_at' missing.\n", i, idInt)
			}
		}

		updatedAtUnix, ok := document["updated_at"].(float64)
		var updatedAt time.Time
		if ok {
			updatedAt = time.Unix(int64(updatedAtUnix), 0)
		} else {
			if val, exists := document["updated_at"]; exists {
				fmt.Printf("WARN: Hit %d (ID: %d): 'updated_at' not float64. Actual type: %T, Value: %+v\n", i, idInt, val, val)
			} else {
				fmt.Printf("WARN: Hit %d (ID: %d): 'updated_at' missing.\n", i, idInt)
			}
		}

		var deletedAt *time.Time
		if deletedAtUnix, ok := document["deleted_at"].(float64); ok && deletedAtUnix > 0 {
			t := time.Unix(int64(deletedAtUnix), 0)
			deletedAt = &t
		} else {
			if val, exists := document["deleted_at"]; exists {
				fmt.Printf("DEBUG: Hit %d (ID: %d): 'deleted_at' not float64 or 0. Actual type: %T, Value: %+v\n", i, idInt, val, val)
			} else {
				fmt.Printf("DEBUG: Hit %d (ID: %d): 'deleted_at' missing or nil.\n", i, idInt)
			}
		}

		exercisesResponse = append(exercisesResponse, dto.ExerciseResponse{
			ID:        idInt, // <--- Assign the converted int ID
			Name:      name,
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
			DeletedAt: deletedAt,
		})
		fmt.Printf("DEBUG: Successfully added doc %d (ID: %d) to response slice.\n", i, idInt)
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
