-- name: CreateUser :one
INSERT INTO users (id, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, email, is_verified, created_at, updated_at;

-- name: GetUserByID :one
SELECT id, email, password_hash, is_verified, created_at, updated_at
FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, is_verified, created_at, updated_at
FROM users
WHERE email = $1 LIMIT 1;

-- name: UpdateUserVerification :exec
UPDATE users
SET is_verified = $2
WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2
WHERE id = $1;