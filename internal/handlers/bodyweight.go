package handlers

import (
	"net/http"
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"time"

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
