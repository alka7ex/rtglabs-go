-- In a new migration file, e.g., db/migrations/20250708_create_logged_exercise_instances.sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS logged_exercise_instances (
    id UUID PRIMARY KEY,
    workout_log_id UUID NOT NULL,       -- Links to the specific workout log this instance belongs to
    exercise_id UUID NOT NULL,          -- Links to the base exercise (e.g., "Push-up", "Squat")
    -- Add any other fields here that represent summary or specific details
    -- about *this particular exercise as it was performed in the log*.
    -- e.g., total_sets_completed INT DEFAULT 0,
    -- e.g., total_reps_completed INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,

    CONSTRAINT fk_logged_exercise_instances_workout_log
        FOREIGN KEY (workout_log_id)
        REFERENCES workout_logs (id)
        ON DELETE CASCADE,

    CONSTRAINT fk_logged_exercise_instances_exercise
        FOREIGN KEY (exercise_id)
        REFERENCES exercises (id)
        ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS logged_exercise_instances;
-- +goose StatementEnd
