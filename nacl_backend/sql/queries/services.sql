-- name: CreateService :one
INSERT INTO services (
    service,
    service_username,
    description,
    encrypted_password,
    nonce,
    encryption_algorithm,
    user_id
    )
VALUES($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetServiceById :one
SELECT *
FROM services
WHERE id = $1;

-- name: GetAllServicesForUserId :many
SELECT * 
FROM services
WHERE user_id = $1

-- name: UpdateService :exec
UPDATE services
SET service_username = $1,
    description = $2,
    encrypted_password = $3,
    nonce = $4,
    encryption_algorithm = $5,
    updated_at = NOW()
WHERE id = $6 AND user_id = $7;

-- name: DeleteServiceById :exec
DELETE FROM services
WHERE id = $1;
