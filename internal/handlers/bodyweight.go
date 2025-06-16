package handlers

import (
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"time"

	bodyweight "rtglabs-go/ent/bodyweight" // <- import this package

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

	query := h.Client.Bodyweight.
		Query().
		Where(bodyweight.DeletedAtIsNil()).
		Order(bodyweight.ByCreatedAt(sql.OrderDesc()))

	if userID != "" {
		uid, err := uuid.Parse(userID)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid user_id")
		}
		query = query.Where(bodyweight.UserIDEQ(uid))
	}

	bodyweights, err := query.All(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{
		"status": "success",
		"data":   bodyweights,
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
		SetDeletedAt(now). // ← Not pointer
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
