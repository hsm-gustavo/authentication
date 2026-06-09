-- name: CreateSession :one
INSERT INTO sessions (id, user_id, secret_hash, ip_address, user_agent)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, secret_hash, ip_address, user_agent, last_verified_at, created_at;

-- name: GetSession :one
SELECT id, user_id, secret_hash, ip_address, user_agent, last_verified_at, created_at
FROM sessions
WHERE id = $1 LIMIT 1;

-- name: UpdateSessionActivity :exec
-- atualiza o last_verified_at e permite atualizar IP/Browser caso mudem sutilmente
UPDATE sessions
SET last_verified_at = NOW(),
    ip_address = $2,
    user_agent = $3
WHERE id = $1;

-- name: DeleteSession :exec
-- usado no Logout
DELETE FROM sessions
WHERE id = $1;

-- name: DeleteAllUserSessionsExceptCurrent :exec
-- desconectar de todos os outros dispositivos
DELETE FROM sessions
WHERE user_id = $1 AND id <> $2;