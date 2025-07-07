package handlers

import (
	"net/http"
	"rtglabs-go/dto"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h *ExerciseHandler) StoreExercise(c echo.Context) error {
	var req dto.CreateExerciseRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	if len(req.Exercises) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "No exercises provided")
	}

	ctx := c.Request().Context()
	now := time.Now()
	builder := h.sq.Insert("exercises").
		Columns("id", "name", "created_at", "updated_at")

	exerciseIDs := make([]uuid.UUID, 0, len(req.Exercises))
	for _, ex := range req.Exercises {
		id := uuid.New()
		exerciseIDs = append(exerciseIDs, id)
		builder = builder.Values(id, ex.Name, now, now)
	}

	query, args, err := builder.ToSql()
	if err != nil {
		c.Logger().Errorf("StoreExercise: Failed to build insert query: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create exercises")
	}

	_, err = h.DB.ExecContext(ctx, query, args...)
	if err != nil {
		c.Logger().Errorf("StoreExercise: Failed to execute insert: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create exercises")
	}

	// Build response DTOs
	exerciseResponses := make([]dto.ExerciseResponse, 0, len(req.Exercises))
	for i, id := range exerciseIDs {
		exerciseResponses = append(exerciseResponses, dto.ExerciseResponse{
			ID:        id,
			Name:      req.Exercises[i].Name,
			CreatedAt: now,
			UpdatedAt: now,
			DeletedAt: nil,
		})
	}

	return c.JSON(http.StatusCreated, dto.CreateExerciseResponse{
		Message:   "Exercises created successfully",
		Exercises: exerciseResponses,
	})
}
