-- name: CreateOperation :exec
INSERT INTO operations (
    user_id,
    op_type,
    service
)
VALUES ($1, $2, $3);

-- name: GetOperationsForUserId :many
SELECT *
FROM operations
WHERE user_id = $1;

-- name: GetOperationsForService :many
SELECT *
FROM operations
WHERE service = $1 AND user_id = $2;

