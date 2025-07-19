-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS profiles (
    id UUID UNIQUE PRIMARY KEY,
    user_id UUID UNIQUE NOT NULL, 
    units INTEGER NOT NULL,
    age INTEGER NOT NULL,
    height NUMERIC(10, 2) NULL,  
    gender INTEGER NULL,
    weight NUMERIC(10, 2) NULL,  
    avatar_url VARCHAR(255) NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL, 

    CONSTRAINT fk_profiles_user
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE 
);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS profiles;
-- +goose StatementEnd
