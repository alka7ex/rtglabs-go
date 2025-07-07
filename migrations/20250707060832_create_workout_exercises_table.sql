-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS workout_exercises (
    id UUID PRIMARY KEY,
    workout_order INTEGER NULL, -- "order" is a keyword in SQL, so it's quoted. Use BIGINT if uint can be very large.
    sets INTEGER NULL,
    weight REAL NULL,        -- Or NUMERIC(8,2), DECIMAL(8,2) if precise decimal required
    reps INTEGER NULL,
    workout_id UUID NOT NULL,          -- REMOVED UNIQUE
    exercise_id UUID NOT NULL,         -- REMOVED UNIQUE
    exercise_instance_id UUID NULL,    -- REMOVED UNIQUE
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,

    CONSTRAINT fk_workout_exercises_workout
        FOREIGN KEY (workout_id)
        REFERENCES workouts (id)
        ON DELETE CASCADE,

    CONSTRAINT fk_workout_exercises_exercise
        FOREIGN KEY (exercise_id)
        REFERENCES exercises (id)
        ON DELETE CASCADE,

    CONSTRAINT fk_workout_exercises_exercise_instance
        FOREIGN KEY (exercise_instance_id)
        REFERENCES exercise_instances (id)
        ON DELETE SET NULL -- Set to NULL if exercise_instance is deleted and this FK is not NULL
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS workout_exercises;
-- +goose StatementEnd
