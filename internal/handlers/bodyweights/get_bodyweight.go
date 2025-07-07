package handlers

import (
	"database/sql" // Import for sql.DB, sql.Null* types, sql.ErrNoRows
	"errors"       // Import errors for errors.Is
	"net/http"

	"rtglabs-go/model" // Import your model package

	"github.com/Masterminds/squirrel" // Import squirrel
	"github.com/google/uuid"
	"github.com/labstack/echo/v4" // Import echo for context, logger, HTTP errors
)

// GetBodyweight retrieves a single bodyweight record by ID (maps to MVC 'Get' or 'Show').
func (h *BodyweightHandler) GetBodyweight(c echo.Context) error {
	// 1. Get the authenticated user ID from the context.
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	// 2. Parse the ID from the URL parameter.
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID format")
	}

	ctx := c.Request().Context()
	var bodyweightRecord model.Bodyweight // Will store the fetched data
	var nullDeletedAt sql.NullTime        // For scanning nullable deleted_at

	// 3. Query the bodyweight record using squirrel, ensuring it belongs to the authenticated user.
	selectQuery, selectArgs, err := h.sq.Select(
		"id", "user_id", "weight", "unit", "created_at", "updated_at", "deleted_at",
	).
		From("bodyweights"). // Table name
		Where(
			squirrel.Eq{"id": id},          // Filter by the record ID
			squirrel.Eq{"user_id": userID}, // Ensure user owns the record
			squirrel.Eq{"deleted_at": nil}, // Only non-deleted records
		).
		ToSql()
	if err != nil {
		c.Logger().Errorf("GetBodyweight: Failed to build select query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve record")
	}

	row := h.DB.QueryRowContext(ctx, selectQuery, selectArgs...)

	err = row.Scan(
		&bodyweightRecord.ID,
		&bodyweightRecord.UserID,
		&bodyweightRecord.Weight,
		&bodyweightRecord.Unit,
		&bodyweightRecord.CreatedAt,
		&bodyweightRecord.UpdatedAt,
		&nullDeletedAt, // Scan into sql.NullTime
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// If no rows found, it means the record doesn't exist, doesn't belong to the user, or is soft-deleted.
			return echo.NewHTTPError(http.StatusNotFound, "Bodyweight not found or you don't have access")
		}
		c.Logger().Errorf("GetBodyweight: Database scan error: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve record")
	}

	// Assign DeletedAt from sql.NullTime to *time.Time in model
	if nullDeletedAt.Valid {
		bodyweightRecord.DeletedAt = &nullDeletedAt.Time
	} else {
		bodyweightRecord.DeletedAt = nil
	}

	// 4. Return the DTO response using the fetched model.Bodyweight.
	return c.JSON(http.StatusOK, toBodyweightResponse(&bodyweightRecord))
}
