-- +goose Up
CREATE TABLE operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    type TEXT NOT NULL,
    service_id UUID NOT NULL,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_service FOREIGN KEY (service_id)
    REFERENCES services(id)
);

-- +goose Down
DROP TABLE operations;
