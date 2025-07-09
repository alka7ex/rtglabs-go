-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS profiles (
    id UUID UNIQUE PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL, -- Foreign Key to users.id, and must be unique for 1:1 relationship
    units INTEGER NOT NULL,
    age INTEGER NOT NULL,
    height NUMERIC(10, 2) NOT NULL, -- Use DECIMAL(10, 2) for MySQL
    gender INTEGER NOT NULL,
    weight NUMERIC(10, 2) NOT NULL, -- Use DECIMAL(10, 2) for MySQL
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL, -- For soft deletes

    CONSTRAINT fk_profiles_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE -- If a user is deleted, their profile is also deleted
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS profiles;
-- +goose StatementEnd
