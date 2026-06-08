-- name: CreateService :one
INSERT INTO services (
    service,
    encrypted_service_username,
    description,
    encrypted_password,
    encryption_algorithm,
    user_id
    )
VALUES($1, $2, $3, $4, $5, $6)
RETURNING id, service, description, encryption_algorithm;

-- name: GetServiceById :one
SELECT *
FROM services
WHERE id = $1;

-- name: GetAllServicesForUserId :many
SELECT id, service, description, encryption_algorithm 
FROM services
WHERE user_id = $1;

-- name: UpdateService :exec
UPDATE services
SET encrypted_service_username = $1,
    description = $2,
    encrypted_password = $3,
    encryption_algorithm = $4,
    updated_at = NOW()
WHERE id = $5 AND user_id = $6;

-- name: DeleteServiceById :exec
DELETE FROM services
WHERE id = $1 AND user_id = $2;
