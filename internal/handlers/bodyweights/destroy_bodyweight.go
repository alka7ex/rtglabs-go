package handlers

import (
	"net/http"
	"time"

	"rtglabs-go/dto" // Your DTOs

	"github.com/Masterminds/squirrel" // Import squirrel
	"github.com/google/uuid"
	"github.com/labstack/echo/v4" // Import echo for context, logger, HTTP errors
)

// Assume BodyweightHandler struct and NewBodyweightHandler are defined correctly
// (using *sql.DB and squirrel.StatementBuilderType, with squirrel.Dollar format)

// DestroyBodyweight performs a soft delete on a bodyweight record.
func (h *BodyweightHandler) DestroyBodyweight(c echo.Context) error {
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
	now := time.Now() // Timestamp for soft delete

	// 3. Build and execute the SQL UPDATE query for soft deletion using squirrel.
	// We set 'deleted_at' to the current time, ensuring it belongs to the user
	// and hasn't been soft-deleted already.
	updateQuery, updateArgs, err := h.sq.Update("bodyweights"). // Table name
									Set("deleted_at", now).
									Where(
			squirrel.Eq{"id": id},          // Filter by the record ID
			squirrel.Eq{"user_id": userID}, // Ensure user owns the record
			squirrel.Eq{"deleted_at": nil}, // Crucial: only update if not already soft-deleted
		).
		ToSql()
	if err != nil {
		c.Logger().Errorf("DestroyBodyweight: Failed to build update query for soft delete: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete record")
	}

	res, err := h.DB.ExecContext(ctx, updateQuery, updateArgs...)
	if err != nil {
		c.Logger().Errorf("DestroyBodyweight: Failed to execute soft delete query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete record")
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		c.Logger().Errorf("DestroyBodyweight: Failed to get rows affected by soft delete: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete record")
	}

	// 4. Check rowsAffected to determine the appropriate response.
	if rowsAffected == 0 {
		// If 0 rows were affected, it means:
		// a) The bodyweight with the given ID was not found for the current user.
		// b) The bodyweight was found for the current user, but it was already soft-deleted
		//    (because of squirrel.Eq{"deleted_at": nil} in the Where clause).
		return c.JSON(http.StatusNotFound, dto.DeleteBodyweightResponse{
			Message: "Bodyweight record not found, not accessible, or already soft-deleted.",
		})
	}

	// 5. If rowsAffected > 0, it means the record was successfully soft-deleted.
	response := dto.DeleteBodyweightResponse{
		Message: "Bodyweight record deleted successfully.",
	}

	return c.JSON(http.StatusOK, response)
}
