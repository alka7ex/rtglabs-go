-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS workout_logs (
    id UUID PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL,      -- Foreign Key to users.id, UNIQUE due to Ent schema
    workout_id UUID UNIQUE NULL,       -- Foreign Key to workouts.id, UNIQUE and NULLABLE
    started_at TIMESTAMP WITH TIME ZONE NULL,
    f_inished_at TIMESTAMP WITH TIME ZONE NULL,
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

CREATE INDEX IF NOT EXISTS idx_workout_logs_status ON workout_logs (status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_workout_logs_user_id ON workout_logs (user_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_workout_logs_workout_id ON workout_logs (workout_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS workout_logs;
-- +goose StatementEnd
