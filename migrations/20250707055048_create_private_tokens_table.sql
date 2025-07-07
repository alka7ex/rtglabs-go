-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS private_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- For PostgreSQL, or UUID() for MySQL, etc.
    user_id UUID  NOT NULL,      -- Foreign Key to users.id, UNIQUE due to Ent schema
    token VARCHAR(255) UNIQUE NOT NULL,
    "type" VARCHAR(255) NOT NULL,       -- CAUTION: Quoting "type" as it's a SQL keyword. Consider renaming.
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_private_tokens_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE -- If user is deleted, delete their private token
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS private_tokens;
-- +goose StatementEnd
