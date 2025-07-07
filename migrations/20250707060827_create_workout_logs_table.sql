-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS workout_logs (
    id UUID PRIMARY KEY,
    user_id UUID  NOT NULL,      -- Foreign Key to users.id, UNIQUE due to Ent schema
    workout_id UUID  NULL,       -- Foreign Key to workouts.id, UNIQUE and NULLABLE
    started_at TIMESTAMP WITH TIME ZONE NULL,
    finished_at TIMESTAMP WITH TIME ZONE NULL,
    status INTEGER NOT NULL DEFAULT 0,
    total_active_duration_seconds BIGINT NOT NULL DEFAULT 0, -- Use BIGINT for uint if it can exceed int32 max
    total_pause_duration_seconds BIGINT NOT NULL DEFAULT 0,  -- Use BIGINT for uint if it can exceed int32 max
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,

    CONSTRAINT fk_workout_logs_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE,

    CONSTRAINT fk_workout_logs_workout
        FOREIGN KEY (workout_id)
        REFERENCES workouts (id)
        ON DELETE CASCADE -- Note: This will delete WorkoutLog if Workout is deleted AND workout_id is not NULL.
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS workout_logs;
-- +goose StatementEnd
