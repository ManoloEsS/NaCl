-- name: CreateOperation :one
INSERT INTO operations (
    user_id,
    type,
    service,
    service_id,
    description
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetOperationsForUserId :many
SELECT *
FROM operations
WHERE user_id = $1;

-- name: GetOperationsForService :many
SELECT *
FROM operations
WHERE service = $1 AND user_id = $2;

-- name: UpdateOperationDesc :exec
UPDATE operations
SET description = $1
WHERE id = $2 AND user_id = $3;
