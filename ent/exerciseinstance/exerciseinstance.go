// Code generated by ent, DO NOT EDIT.

package exerciseinstance

import (
	"time"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"github.com/google/uuid"
)

const (
	// Label holds the string label denoting the exerciseinstance type in the database.
	Label = "exercise_instance"
	// FieldID holds the string denoting the id field in the database.
	FieldID = "id"
	// FieldCreatedAt holds the string denoting the created_at field in the database.
	FieldCreatedAt = "created_at"
	// FieldUpdatedAt holds the string denoting the updated_at field in the database.
	FieldUpdatedAt = "updated_at"
	// FieldDeletedAt holds the string denoting the deleted_at field in the database.
	FieldDeletedAt = "deleted_at"
	// EdgeExercise holds the string denoting the exercise edge name in mutations.
	EdgeExercise = "exercise"
	// EdgeWorkoutExercises holds the string denoting the workout_exercises edge name in mutations.
	EdgeWorkoutExercises = "workout_exercises"
	// EdgeExerciseSets holds the string denoting the exercise_sets edge name in mutations.
	EdgeExerciseSets = "exercise_sets"
	// EdgeWorkoutLog holds the string denoting the workout_log edge name in mutations.
	EdgeWorkoutLog = "workout_log"
	// Table holds the table name of the exerciseinstance in the database.
	Table = "exercise_instances"
	// ExerciseTable is the table that holds the exercise relation/edge.
	ExerciseTable = "exercise_instances"
	// ExerciseInverseTable is the table name for the Exercise entity.
	// It exists in this package in order to avoid circular dependency with the "exercise" package.
	ExerciseInverseTable = "exercises"
	// ExerciseColumn is the table column denoting the exercise relation/edge.
	ExerciseColumn = "exercise_exercise_instances"
	// WorkoutExercisesTable is the table that holds the workout_exercises relation/edge.
	WorkoutExercisesTable = "workout_exercises"
	// WorkoutExercisesInverseTable is the table name for the WorkoutExercise entity.
	// It exists in this package in order to avoid circular dependency with the "workoutexercise" package.
	WorkoutExercisesInverseTable = "workout_exercises"
	// WorkoutExercisesColumn is the table column denoting the workout_exercises relation/edge.
	WorkoutExercisesColumn = "exercise_instance_workout_exercises"
	// ExerciseSetsTable is the table that holds the exercise_sets relation/edge.
	ExerciseSetsTable = "exercise_sets"
	// ExerciseSetsInverseTable is the table name for the ExerciseSet entity.
	// It exists in this package in order to avoid circular dependency with the "exerciseset" package.
	ExerciseSetsInverseTable = "exercise_sets"
	// ExerciseSetsColumn is the table column denoting the exercise_sets relation/edge.
	ExerciseSetsColumn = "exercise_instance_exercise_sets"
	// WorkoutLogTable is the table that holds the workout_log relation/edge.
	WorkoutLogTable = "exercise_instances"
	// WorkoutLogInverseTable is the table name for the WorkoutLog entity.
	// It exists in this package in order to avoid circular dependency with the "workoutlog" package.
	WorkoutLogInverseTable = "workout_logs"
	// WorkoutLogColumn is the table column denoting the workout_log relation/edge.
	WorkoutLogColumn = "workout_log_exercise_instances"
)

// Columns holds all SQL columns for exerciseinstance fields.
var Columns = []string{
	FieldID,
	FieldCreatedAt,
	FieldUpdatedAt,
	FieldDeletedAt,
}

// ForeignKeys holds the SQL foreign-keys that are owned by the "exercise_instances"
// table and are not defined as standalone fields in the schema.
var ForeignKeys = []string{
	"exercise_exercise_instances",
	"workout_log_exercise_instances",
}

// ValidColumn reports if the column name is valid (part of the table columns).
func ValidColumn(column string) bool {
	for i := range Columns {
		if column == Columns[i] {
			return true
		}
	}
	for i := range ForeignKeys {
		if column == ForeignKeys[i] {
			return true
		}
	}
	return false
}

var (
	// DefaultCreatedAt holds the default value on creation for the "created_at" field.
	DefaultCreatedAt func() time.Time
	// DefaultUpdatedAt holds the default value on creation for the "updated_at" field.
	DefaultUpdatedAt func() time.Time
	// UpdateDefaultUpdatedAt holds the default value on update for the "updated_at" field.
	UpdateDefaultUpdatedAt func() time.Time
	// DefaultID holds the default value on creation for the "id" field.
	DefaultID func() uuid.UUID
)

// OrderOption defines the ordering options for the ExerciseInstance queries.
type OrderOption func(*sql.Selector)

// ByID orders the results by the id field.
func ByID(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldID, opts...).ToFunc()
}

// ByCreatedAt orders the results by the created_at field.
func ByCreatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldCreatedAt, opts...).ToFunc()
}

// ByUpdatedAt orders the results by the updated_at field.
func ByUpdatedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldUpdatedAt, opts...).ToFunc()
}

// ByDeletedAt orders the results by the deleted_at field.
func ByDeletedAt(opts ...sql.OrderTermOption) OrderOption {
	return sql.OrderByField(FieldDeletedAt, opts...).ToFunc()
}

// ByExerciseField orders the results by exercise field.
func ByExerciseField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newExerciseStep(), sql.OrderByField(field, opts...))
	}
}

// ByWorkoutExercisesCount orders the results by workout_exercises count.
func ByWorkoutExercisesCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newWorkoutExercisesStep(), opts...)
	}
}

// ByWorkoutExercises orders the results by workout_exercises terms.
func ByWorkoutExercises(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newWorkoutExercisesStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByExerciseSetsCount orders the results by exercise_sets count.
func ByExerciseSetsCount(opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborsCount(s, newExerciseSetsStep(), opts...)
	}
}

// ByExerciseSets orders the results by exercise_sets terms.
func ByExerciseSets(term sql.OrderTerm, terms ...sql.OrderTerm) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newExerciseSetsStep(), append([]sql.OrderTerm{term}, terms...)...)
	}
}

// ByWorkoutLogField orders the results by workout_log field.
func ByWorkoutLogField(field string, opts ...sql.OrderTermOption) OrderOption {
	return func(s *sql.Selector) {
		sqlgraph.OrderByNeighborTerms(s, newWorkoutLogStep(), sql.OrderByField(field, opts...))
	}
}
func newExerciseStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ExerciseInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, ExerciseTable, ExerciseColumn),
	)
}
func newWorkoutExercisesStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(WorkoutExercisesInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, WorkoutExercisesTable, WorkoutExercisesColumn),
	)
}
func newExerciseSetsStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(ExerciseSetsInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.O2M, false, ExerciseSetsTable, ExerciseSetsColumn),
	)
}
func newWorkoutLogStep() *sqlgraph.Step {
	return sqlgraph.NewStep(
		sqlgraph.From(Table, FieldID),
		sqlgraph.To(WorkoutLogInverseTable, FieldID),
		sqlgraph.Edge(sqlgraph.M2O, true, WorkoutLogTable, WorkoutLogColumn),
	)
}
