package handler

import (
	"rtglabs-go/ent"
)

type WorkoutLogHandler struct {
	Client *ent.Client
}

func NewWorkoutLogHandler(client *ent.Client) *WorkoutHandler {
	return &WorkoutHandler{Client: client}
}
