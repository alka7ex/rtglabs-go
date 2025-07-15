package handlers

import (
	"database/sql" // Import for sql.DB, sql.Null* types, sql.ErrNoRows
	"errors"       // Import errors for errors.Is
	"net/http"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/model" // Import your model package

	"github.com/Masterminds/squirrel" // Import squirrel
	"github.com/google/uuid"
	"github.com/labstack/echo/v4" // Import echo for context, logger, HTTP errors
)

// StoreBodyweight creates a new bodyweight record (maps to MVC 'Store' or 'Post').
func (h *BodyweightHandler) StoreBodyweight(c echo.Context) error {
	// 1. Get the authenticated user ID from the context.
	userID, ok := c.Get("user_id").(uuid.UUID)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, "User ID not found in context")
	}

	// 2. Bind and validate the request body.
	var req dto.CreateBodyweightRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}
	// Assuming `c.Validate` is an Echo middleware or helper.
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()
	newBodyweightID := uuid.New() // Generate a new UUID for the bodyweight record
	currentTime := time.Now()     // Get current time for CreatedAt and UpdatedAt

	// 3. Insert the new bodyweight record into the database using squirrel.
	// Removed "unit" column and unitString from values.
	insertQuery, insertArgs, err := h.sq.Insert("bodyweights"). // Table name
									Columns("id", "user_id", "weight", "created_at", "updated_at").        // Columns to insert
									Values(newBodyweightID, userID, req.Weight, currentTime, currentTime). // Values
									ToSql()
	if err != nil {
		c.Logger().Errorf("StoreBodyweight: Failed to build insert query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create bodyweight record")
	}

	_, err = h.DB.ExecContext(ctx, insertQuery, insertArgs...)
	if err != nil {
		// Log the database error. Add more specific error handling if needed (e.g., for constraint violations).
		c.Logger().Errorf("StoreBodyweight: Failed to insert new bodyweight record: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create bodyweight record")
	}

	// 4. Fetch the newly created record to populate the response DTO.
	// We need to fetch it to get the confirmed timestamps (if DB handles them)
	// and ensure data consistency.
	var createdBodyweight model.Bodyweight
	var nullDeletedAt sql.NullTime // Use sql.NullTime for the nullable DeletedAt field

	// Removed "unit" from the SELECT columns
	fetchQuery, fetchArgs, err := h.sq.Select("id", "user_id", "weight", "created_at", "updated_at", "deleted_at").
		From("bodyweights").
		Where(squirrel.Eq{"id": newBodyweightID}). // Fetch by the ID we just inserted
		ToSql()
	if err != nil {
		c.Logger().Errorf("StoreBodyweight: Failed to build fetch query for new record: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve created bodyweight record")
	}

	row := h.DB.QueryRowContext(ctx, fetchQuery, fetchArgs...)

	// Removed &createdBodyweight.Unit from the Scan arguments
	err = row.Scan(
		&createdBodyweight.ID,
		&createdBodyweight.UserID,
		&createdBodyweight.Weight,
		&createdBodyweight.CreatedAt,
		&createdBodyweight.UpdatedAt,
		&nullDeletedAt, // Scan into sql.NullTime
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// This case is unlikely if the insert just succeeded, but handle defensively.
			c.Logger().Errorf("StoreBodyweight: Newly created bodyweight record not found after insert: %v", err)
			return echo.NewHTTPError(http.StatusNotFound, "Created bodyweight record not found after insertion")
		}
		c.Logger().Errorf("StoreBodyweight: Database scan error for created bodyweight: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve created bodyweight record details")
	}

	// Assign DeletedAt from sql.NullTime to *time.Time in model
	if nullDeletedAt.Valid {
		createdBodyweight.DeletedAt = &nullDeletedAt.Time
	} else {
		createdBodyweight.DeletedAt = nil
	}

	// 5. Build the DTO response using the fetched model.Bodyweight.
	response := dto.CreateBodyweightResponse{
		Message:    "Bodyweight record created successfully.",
		Bodyweight: toBodyweightResponse(&createdBodyweight), // Use the fetched 'createdBodyweight'
	}

	return c.JSON(http.StatusCreated, response)
}

