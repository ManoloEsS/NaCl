-- name: CreateCredential :one
INSERT INTO credentials (
    service,
    encrypted_service_username,
    description,
    encrypted_password,
    encryption_algorithm,
    user_id
    )
VALUES($1, $2, $3, $4, $5, $6)
RETURNING id, service, description, encryption_algorithm;

-- name: GetCredentialById :one
SELECT *
FROM credentials
WHERE id = $1;

-- name: GetAllCredentialsForUserId :many
SELECT id, service, description, encryption_algorithm 
FROM credentials
WHERE user_id = $1;

-- name: UpdateCredential :one
UPDATE credentials
SET encrypted_password = $1,
    encryption_algorithm = $2,
    updated_at = NOW()
WHERE id = $3 AND user_id = $4
RETURNING id, service, description, encryption_algorithm;

-- name: DeleteCredentialById :one
DELETE FROM credentials
WHERE id = $1
RETURNING service;
