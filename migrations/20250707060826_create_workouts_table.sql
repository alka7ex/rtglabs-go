-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS workouts (
    id UUID PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL, -- Foreign Key to users.id, and must be unique for 1:1 relationship
    name VARCHAR(255) NOT NULL,    -- Assuming NOT NULL as field.String defaults to Required in Ent
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,

    CONSTRAINT fk_workouts_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE -- If user is deleted, delete their workout
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_workouts_user_id ON workouts (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS workouts;
-- +goose StatementEnd
