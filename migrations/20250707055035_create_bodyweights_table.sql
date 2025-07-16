-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS bodyweights (
    id UUID UNIQUE PRIMARY KEY,
    user_id UUID NOT NULL,      -- Foreign Key to users.id, UNIQUE due to Ent schema
    weight REAL NOT NULL,               -- Use NUMERIC(10, 2) or DECIMAL(10, 2) if higher precision is needed
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,

    CONSTRAINT fk_bodyweights_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE -- If user is deleted, delete their bodyweight record
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bodyweights;
-- +goose StatementEnd
