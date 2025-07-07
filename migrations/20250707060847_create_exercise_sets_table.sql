-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS exercise_sets (
    id UUID PRIMARY KEY,
    weight NUMERIC(8,2) NULL, -- Or DECIMAL(8,2) for MySQL/other DBs
    reps INTEGER NULL,
    set_number INTEGER NOT NULL,
    finished_at TIMESTAMP WITH TIME ZONE NULL,
    status INTEGER NOT NULL DEFAULT 0,
    workout_log_id UUID  NOT NULL,       -- Foreign Key to workout_logs.id, UNIQUE due to Ent schema
    exercise_id UUID  NOT NULL,          -- Foreign Key to exercises.id, UNIQUE due to Ent schema
    exercise_instance_id UUID  NULL,     -- Foreign Key to exercise_instances.id, UNIQUE and NULLABLE
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

    CONSTRAINT fk_exercise_sets_exercise_instance
        FOREIGN KEY (exercise_instance_id)
        REFERENCES exercise_instances (id)
        ON DELETE CASCADE -- If exercise_instance is deleted, delete this set (if FK is not NULL)
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS exercise_sets;
-- +goose StatementEnd
