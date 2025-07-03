package handler

import (
	"rtglabs-go/dto"
	"rtglabs-go/ent"
	"time"

	"github.com/google/uuid"
)

type WorkoutHandler struct {
	Client *ent.Client
}

func NewWorkoutHandler(client *ent.Client) *WorkoutHandler {
	return &WorkoutHandler{Client: client}
}

// DestroyWorkout performs a soft delete on a workout and its associated workout exercises.
func toWorkoutResponse(w *ent.Workout) dto.WorkoutResponse {
	var deletedAt *time.Time
	if w.DeletedAt != nil {
		deletedAt = w.DeletedAt
	}

	var exercises []dto.WorkoutExerciseResponse
	for _, we := range w.Edges.WorkoutExercises {
		exercises = append(exercises, toWorkoutExerciseResponse(we))
	}

	var userID uuid.UUID
	if w.Edges.User != nil {
		userID = w.Edges.User.ID
	}

	return dto.WorkoutResponse{
		ID:               w.ID,
		UserID:           userID,
		Name:             w.Name,
		CreatedAt:        w.CreatedAt,
		UpdatedAt:        w.UpdatedAt,
		DeletedAt:        deletedAt,
		WorkoutExercises: exercises,
	}
}

func toWorkoutExerciseResponse(we *ent.WorkoutExercise) dto.WorkoutExerciseResponse {
	var deletedAt *time.Time
	if we.DeletedAt != nil {
		deletedAt = we.DeletedAt
	}

	var instanceID *uuid.UUID
	if we.Edges.ExerciseInstance != nil {
		instanceID = &we.Edges.ExerciseInstance.ID
	}

	var exerciseDTO *dto.ExerciseResponse
	if we.Edges.Exercise != nil {
		ex := toExerciseResponse(we.Edges.Exercise)
		exerciseDTO = &ex
	}

	var instanceDTO *dto.ExerciseInstanceResponse
	if we.Edges.ExerciseInstance != nil {
		inst := toExerciseInstanceResponse(we.Edges.ExerciseInstance)
		instanceDTO = &inst
	}

	var workoutID uuid.UUID
	if we.Edges.Workout != nil {
		workoutID = we.Edges.Workout.ID
	}

	var exerciseID uuid.UUID
	if we.Edges.Exercise != nil {
		exerciseID = we.Edges.Exercise.ID
	}

	return dto.WorkoutExerciseResponse{
		ID:                 we.ID,
		WorkoutID:          workoutID,
		ExerciseID:         exerciseID,
		ExerciseInstanceID: instanceID,
		Order:              we.Order,
		Sets:               we.Sets,
		Weight:             we.Weight,
		Reps:               we.Reps,
		CreatedAt:          we.CreatedAt,
		UpdatedAt:          we.UpdatedAt,
		DeletedAt:          deletedAt,
		Exercise:           exerciseDTO,
		ExerciseInstance:   instanceDTO,
	}
}

func toExerciseResponse(ex *ent.Exercise) dto.ExerciseResponse {
	var deletedAt *time.Time
	if ex.DeletedAt != nil {
		deletedAt = ex.DeletedAt
	}
	return dto.ExerciseResponse{
		ID:        ex.ID,
		Name:      ex.Name,
		CreatedAt: ex.CreatedAt,
		UpdatedAt: ex.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

func toExerciseInstanceResponse(ei *ent.ExerciseInstance) dto.ExerciseInstanceResponse {
	var workoutLogID *uuid.UUID
	if ei.Edges.WorkoutLog != nil {
		workoutLogID = &ei.Edges.WorkoutLog.ID
	}

	var deletedAt *time.Time
	if ei.DeletedAt != nil {
		deletedAt = ei.DeletedAt
	}

	var exerciseID uuid.UUID
	if ei.Edges.Exercise != nil {
		exerciseID = ei.Edges.Exercise.ID
	}

	return dto.ExerciseInstanceResponse{
		ID:           ei.ID,
		WorkoutLogID: workoutLogID,
		ExerciseID:   exerciseID,
		CreatedAt:    ei.CreatedAt,
		UpdatedAt:    ei.UpdatedAt,
		DeletedAt:    deletedAt,
	}
}
