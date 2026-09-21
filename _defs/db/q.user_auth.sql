-- name: StoreAccessToken :exec
INSERT INTO access_tokens (token, expires_at) VALUES (?, ?);

-- name: GetAccessToken :one
SELECT * FROM access_tokens WHERE token = ? and is_active = 1;

-- name: DeleteAccessToken :exec
UPDATE  access_tokens SET is_active = 0 WHERE token = ? AND is_active = 1;

-- name: DeleteExpiredAccessTokens :exec
DELETE FROM access_tokens WHERE expires_at < CURRENT_TIMESTAMP OR is_active = 0;
