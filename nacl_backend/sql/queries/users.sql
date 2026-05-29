-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: GetUserById :one
SELECT * FROM users WHERE id = $1;

-- name: CreateUser :one
INSERT INTO users (username, password_hash)
VALUES (
    $1,
    $2
    )
    RETURNING *;

-- name: DeleteUser :one
DELETE FROM users
WHERE id = $1;

-- name: UpdateUserPasswordHash :one
UPDATE users
SET password_hash = $1, updated_at = NOW()
WHERE id = $2
RETURNING *;

-- name: UpdateUserSalt :exec
UPDATE users
SET master_key_salt = $1
WHERE id = $2;

-- name: UpdateUserKey :exec
UPDATE users
SET encrypted_master_key = $1
WHERE id = $2;
