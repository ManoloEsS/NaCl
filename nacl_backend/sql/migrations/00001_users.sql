-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    username TEXT NOT NULL UNIQUE,
    password_hash VARCHAR(512) NOT NULL,
    master_key_salt VARCHAR(255) NOT NULL,
    encrypted_master_key TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE users;
