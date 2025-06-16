package dto

import "github.com/google/uuid"

type CreateBodyweightRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Weight float64   `json:"weight" validate:"required,gt=0"`
	Unit   string    `json:"unit" validate:"required"`
}
