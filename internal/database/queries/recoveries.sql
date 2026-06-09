-- name: CreateRecovery :one
INSERT INTO recoveries (user_id, email, code, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, email, code, attempts, expired, expires_at, created_at, updated_at;

-- name: GetActiveRecoveryByCode :one
-- busca apenas códigos que pertencem ao e-mail, não estão expirados, não foram usados e têm menos de 5 tentativas
SELECT id, user_id, email, code, attempts, expired, expires_at, created_at, updated_at
FROM recoveries
WHERE code = $1 
  AND email = $2 
  AND expired = FALSE 
  AND expires_at > NOW() 
  AND attempts < 5 
LIMIT 1;

-- name: IncrementRecoveryAttempts :exec
-- registra um erro de digitação do código
UPDATE recoveries
SET attempts = attempts + 1
WHERE id = $1;

-- name: MarkRecoveryAsUsed :exec
-- invalida o código imediatamente após o sucesso
UPDATE recoveries
SET expired = TRUE
WHERE id = $1;