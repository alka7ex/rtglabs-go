-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS exercise_instances (
    id UUID PRIMARY KEY,
    exercise_id UUID NOT NULL,               -- REMOVED UNIQUE HERE
    workout_log_id UUID  NULL,         -- This UNIQUE is fine, as a workout_log likely links to one main instance
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,

    CONSTRAINT fk_exercise_instances_exercise
        FOREIGN KEY (exercise_id)
        REFERENCES exercises (id)
        ON DELETE CASCADE,

    CONSTRAINT fk_exercise_instances_workout_log
        FOREIGN KEY (workout_log_id)
        REFERENCES workout_logs (id)
        ON DELETE SET NULL
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS exercise_instances;
-- +goose StatementEnd
