-- +goose Up

CREATE TABLE services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    service TEXT NOT NULL,
    service_username TEXT NOT NULL,
    description TEXT,
    encrypted_password TEXT NOT NULL,
    nonce TEXT NOT NULL,
    encryption_algorithm TEXT NOT NULL,
    user_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_user FOREIGN KEY (user_id)
    REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down

DROP TABLE services;
