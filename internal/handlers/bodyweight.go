package handlers

import (
	"fmt"
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"strconv" // Import for string to int conversion
	"time"

	bodyweight "rtglabs-go/ent/bodyweight"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type BodyweightHandler struct {
	Client *ent.Client
}

func NewBodyweightHandler(client *ent.Client) *BodyweightHandler {
	return &BodyweightHandler{Client: client}
}

func (h *BodyweightHandler) CreateBodyweight(c echo.Context) error {
	var req dto.CreateBodyweightRequest

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	bw, err := h.Client.Bodyweight.
		Create().
		SetUserID(req.UserID).
		SetWeight(req.Weight).
		SetUnit(req.Unit).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(c.Request().Context())

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"status":  "success",
		"message": "Bodyweight record created",
		"data":    bw,
	})
}

func (h *BodyweightHandler) ListBodyweights(c echo.Context) error {
	userID := c.QueryParam("user_id")

	// --- Pagination Parameters ---
	pageStr := c.QueryParam("page")
	limitStr := c.QueryParam("limit")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1 // Default to first page
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 25 // Default limit per page
	}
	if limit > 100 { // Cap the limit to prevent excessively large requests
		limit = 100
	}

	offset := (page - 1) * limit
	// --- End Pagination Parameters ---

	// Base query builder
	query := h.Client.Bodyweight.
		Query().
		Where(bodyweight.DeletedAtIsNil())

	// Apply UserID filter if present
	if userID != "" {
		uid, err := uuid.Parse(userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user_id format")
		}
		query = query.Where(bodyweight.UserIDEQ(uid))
	}

	// Get total count BEFORE applying limit and offset
	totalCount, err := query.Count(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to count bodyweights: %v", err))
	}

	// Apply sorting, limit, and offset for pagination
	bodyweights, err := query.
		Order(bodyweight.ByCreatedAt(sql.OrderDesc())). // Always order for consistent pagination
		Limit(limit).
		Offset(offset).
		All(c.Request().Context())

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("Failed to list bodyweights: %v", err))
	}

	// Calculate total pages
	totalPages := (totalCount + limit - 1) / limit

	return c.JSON(http.StatusOK, echo.Map{
		"status": "success",
		"data":   bodyweights,
		"pagination": echo.Map{
			"total_items":  totalCount,
			"total_pages":  totalPages,
			"current_page": page,
			"per_page":     limit,
		},
	})
}

func (h *BodyweightHandler) GetBodyweight(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	bw, err := h.Client.Bodyweight.
		Query().
		Where(bodyweight.IDEQ(id), bodyweight.DeletedAtIsNil()).
		Only(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Bodyweight not found")
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status": "success",
		"data":   bw,
	})
}

func (h *BodyweightHandler) UpdateBodyweight(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	var req dto.UpdateBodyweightRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid body")
	}
	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	bw, err := h.Client.Bodyweight.
		UpdateOneID(id).
		SetWeight(req.Weight).
		SetUnit(req.Unit).
		SetUpdatedAt(time.Now()).
		Save(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status":  "success",
		"message": "Bodyweight updated",
		"data":    bw,
	})
}

func (h *BodyweightHandler) DeleteBodyweight(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid ID")
	}

	now := time.Now()

	_, err = h.Client.Bodyweight.
		UpdateOneID(id).
		SetDeletedAt(now).
		Save(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status":  "success",
		"message": "Bodyweight deleted",
	})
}

func timePtr(t time.Time) *time.Time {
	return &t
}

