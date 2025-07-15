package handlers

import (
	"database/sql" // Import for sql.DB, sql.Null* types, sql.ErrNoRows
	"errors"
	"net/http"
	"time"

	"rtglabs-go/dto"
	"rtglabs-go/model" // Import your model package

	"github.com/Masterminds/squirrel" // Import squirrel
	"github.com/google/uuid"
	"github.com/labstack/echo/v4" // Import echo for context, logger, HTTP errors
)

func (h *BodyweightHandler) UpdateBodyweight(c echo.Context) error {
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

	// 3. Bind and validate the request body.
	var req dto.UpdateBodyweightRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ctx := c.Request().Context()
	currentTime := time.Now() // Time for UpdatedAt

	// Removed DTO's Unit conversion and validation check
	// unitString := ""
	// if req.Unit != nil {
	// 	unitString = strconv.Itoa(*req.Unit)
	// } else {
	// 	return echo.NewHTTPError(http.StatusBadRequest, "Unit is required and cannot be nil")
	// }

	// 4. Update the record using squirrel, ensuring it belongs to the authenticated user.
	updateQuery, updateArgs, err := h.sq.Update("bodyweights"). // Table name
									Set("weight", req.Weight).
		// Removed Set("unit", unitString).
		Set("updated_at", currentTime).
		Where(
			squirrel.Eq{"id": id},
			squirrel.Eq{"user_id": userID}, // Ensure user owns the record
			squirrel.Eq{"deleted_at": nil}, // Only update non-deleted records
		).
		ToSql()
	if err != nil {
		c.Logger().Errorf("UpdateBodyweight: Failed to build update query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update bodyweight record")
	}

	res, err := h.DB.ExecContext(ctx, updateQuery, updateArgs...)
	if err != nil {
		c.Logger().Errorf("UpdateBodyweight: Failed to execute update query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update record")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		c.Logger().Errorf("UpdateBodyweight: Failed to get rows affected: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update record")
	}

	// Check if any rows were affected (i.e., the record was found and updated)
	if rowsAffected == 0 {
		// This means either the ID didn't exist, or it didn't belong to the user, or it was already soft-deleted.
		return echo.NewHTTPError(http.StatusNotFound, "Bodyweight not found or you don't have access")
	}

	// 5. Fetch the updated record from the database to return it in the response.
	// This is a separate query, as the SQL UPDATE doesn't return the full updated entity.
	var updatedBodyweight model.Bodyweight
	var nullDeletedAt sql.NullTime // For nullable DeletedAt field

	fetchQuery, fetchArgs, err := h.sq.Select(
		"id", "user_id", "weight", "created_at", "updated_at", "deleted_at", // Removed "unit"
	).
		From("bodyweights").
		Where(
			squirrel.Eq{"id": id},
			squirrel.Eq{"user_id": userID}, // Re-verify ownership
			squirrel.Eq{"deleted_at": nil}, // Ensure we fetch the non-deleted one
		).
		ToSql()
	if err != nil {
		c.Logger().Errorf("UpdateBodyweight: Failed to build fetch query for updated record: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve updated record details")
	}

	row := h.DB.QueryRowContext(ctx, fetchQuery, fetchArgs...)

	err = row.Scan(
		&updatedBodyweight.ID,
		&updatedBodyweight.UserID,
		&updatedBodyweight.Weight,
		// Removed &updatedBodyweight.Unit,
		&updatedBodyweight.CreatedAt,
		&updatedBodyweight.UpdatedAt,
		&nullDeletedAt, // Scan into sql.NullTime
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// This case should ideally not happen if rowsAffected > 0, but can if record was deleted concurrently.
			c.Logger().Errorf("UpdateBodyweight: Updated bodyweight record not found after update: %v", err)
			return echo.NewHTTPError(http.StatusNotFound, "Updated bodyweight record not found after update")
		}
		c.Logger().Errorf("UpdateBodyweight: Database scan error for updated bodyweight: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve updated record details")
	}

	// Assign DeletedAt from sql.NullTime to *time.Time in model
	if nullDeletedAt.Valid {
		updatedBodyweight.DeletedAt = &nullDeletedAt.Time
	} else {
		updatedBodyweight.DeletedAt = nil
	}

	// 6. Build the DTO response and return.
	response := dto.UpdateBodyweightResponse{
		Message:    "Bodyweight record updated successfully.",
		Bodyweight: toBodyweightResponse(&updatedBodyweight), // Use the fetched 'updatedBodyweight'
	}

	return c.JSON(http.StatusOK, response)
}
