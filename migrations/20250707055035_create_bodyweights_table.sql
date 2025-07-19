-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS bodyweights (
    id UUID UNIQUE PRIMARY KEY,
    user_id UUID NOT NULL,
    weight REAL NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,

    CONSTRAINT fk_bodyweights_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE 
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bodyweights;
-- +goose StatementEnd
