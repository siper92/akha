-- name: UpsertWorker :exec
INSERT INTO workers (name, token_hash, tier) VALUES (?, ?, ?)
ON CONFLICT(token_hash) DO UPDATE SET name = excluded.name, tier = excluded.tier, is_active = 1;

-- name: GetWorkerByTokenHash :one
SELECT * FROM workers WHERE token_hash = ? AND is_active = 1;

-- name: StoreIssuedToken :exec
INSERT INTO issued_tokens (worker_id, token, expires_at) VALUES (?, ?, ?);

-- name: GetIssuedToken :one
SELECT * FROM issued_tokens WHERE token = ? AND is_active = 1;

-- name: RevokeIssuedToken :exec
UPDATE issued_tokens SET is_active = 0 WHERE token = ? AND is_active = 1;

-- name: DeleteExpiredIssuedTokens :exec
DELETE FROM issued_tokens WHERE expires_at < CURRENT_TIMESTAMP OR is_active = 0;

-- name: RecordLogin :exec
INSERT INTO login_log (worker_id, addr, ok, reason, created_at) VALUES (?, ?, ?, ?, ?);
