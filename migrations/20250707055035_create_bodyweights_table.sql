-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS bodyweights (
    id UUID PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL,      -- Foreign Key to users.id, UNIQUE due to Ent schema
    weight REAL NOT NULL,               -- Use NUMERIC(10, 2) or DECIMAL(10, 2) if higher precision is needed
    unit VARCHAR(50) NOT NULL,          -- Assuming a reasonable length for unit string
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,

    CONSTRAINT fk_bodyweights_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE -- If user is deleted, delete their bodyweight record
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_bodyweights_user_id ON bodyweights (user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bodyweights;
-- +goose StatementEnd
