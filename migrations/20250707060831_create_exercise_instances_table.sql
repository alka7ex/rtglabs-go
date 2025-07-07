-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS exercise_instances (
    id UUID PRIMARY KEY,
    exercise_id UUID UNIQUE NOT NULL,      -- Foreign Key to exercises.id, UNIQUE due to Ent schema
    workout_log_id UUID UNIQUE NULL,       -- Foreign Key to workout_logs.id, UNIQUE and NULLABLE
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,

    CONSTRAINT fk_exercise_instances_exercise
        FOREIGN KEY (exercise_id)
        REFERENCES exercises (id)
        ON DELETE CASCADE, -- Example: if exercise is deleted, delete its instance

    CONSTRAINT fk_exercise_instances_workout_log
        FOREIGN KEY (workout_log_id)
        REFERENCES workout_logs (id)
        ON DELETE SET NULL -- As per entsql.SetNull annotation
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_exercise_instances_exercise_id ON exercise_instances (exercise_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_exercise_instances_workout_log_id ON exercise_instances (workout_log_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS exercise_instances;
-- +goose StatementEnd
