-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS exercise_sets (
    id UUID PRIMARY KEY,
    weight NUMERIC(8,2) NULL,
    reps INTEGER NULL,
    set_number INTEGER NOT NULL,
    finished_at TIMESTAMP WITH TIME ZONE NULL,
    status INTEGER NOT NULL DEFAULT 0,
    workout_log_id UUID  NOT NULL,
    exercise_id UUID  NOT NULL,
    logged_exercise_instance_id UUID  NULL, -- This is the column in exercise_sets
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,

    CONSTRAINT fk_exercise_sets_workout_log
        FOREIGN KEY (workout_log_id)
        REFERENCES workout_logs (id)
        ON DELETE CASCADE,

    CONSTRAINT fk_exercise_sets_exercise
        FOREIGN KEY (exercise_id)
        REFERENCES exercises (id)
        ON DELETE CASCADE,

    CONSTRAINT fk_exercise_sets_logged_exercise_instance -- New constraint name (good practice)
        FOREIGN KEY (logged_exercise_instance_id)        -- Referencing the correct column in exercise_sets
        REFERENCES logged_exercise_instances (id)      -- Referencing the correct table and its ID column
        ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS exercise_sets;
-- +goose StatementEnd
