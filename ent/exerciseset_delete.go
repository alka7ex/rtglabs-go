// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"rtglabs-go/ent/exerciseset"
	"rtglabs-go/ent/predicate"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// ExerciseSetDelete is the builder for deleting a ExerciseSet entity.
type ExerciseSetDelete struct {
	config
	hooks    []Hook
	mutation *ExerciseSetMutation
}

// Where appends a list predicates to the ExerciseSetDelete builder.
func (esd *ExerciseSetDelete) Where(ps ...predicate.ExerciseSet) *ExerciseSetDelete {
	esd.mutation.Where(ps...)
	return esd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (esd *ExerciseSetDelete) Exec(ctx context.Context) (int, error) {
	return withHooks(ctx, esd.sqlExec, esd.mutation, esd.hooks)
}

// ExecX is like Exec, but panics if an error occurs.
func (esd *ExerciseSetDelete) ExecX(ctx context.Context) int {
	n, err := esd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (esd *ExerciseSetDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := sqlgraph.NewDeleteSpec(exerciseset.Table, sqlgraph.NewFieldSpec(exerciseset.FieldID, field.TypeUUID))
	if ps := esd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, esd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	esd.mutation.done = true
	return affected, err
}

// ExerciseSetDeleteOne is the builder for deleting a single ExerciseSet entity.
type ExerciseSetDeleteOne struct {
	esd *ExerciseSetDelete
}

// Where appends a list predicates to the ExerciseSetDelete builder.
func (esdo *ExerciseSetDeleteOne) Where(ps ...predicate.ExerciseSet) *ExerciseSetDeleteOne {
	esdo.esd.mutation.Where(ps...)
	return esdo
}

// Exec executes the deletion query.
func (esdo *ExerciseSetDeleteOne) Exec(ctx context.Context) error {
	n, err := esdo.esd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{exerciseset.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (esdo *ExerciseSetDeleteOne) ExecX(ctx context.Context) {
	if err := esdo.Exec(ctx); err != nil {
		panic(err)
	}
}
