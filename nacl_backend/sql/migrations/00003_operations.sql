-- +goose Up
CREATE TABLE operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    user_id UUID NOT NULL,
    op_type TEXT NOT NULL,
    service TEXT NOT NULL,
    service_id UUID,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_service FOREIGN KEY (service_id)
    REFERENCES services(id),
    CONSTRAINT fk_user FOREIGN KEY (user_id)
    REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE operations;
