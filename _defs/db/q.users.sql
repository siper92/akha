-- name: getUser :one
SELECT * FROM access_tokens WHERE token = ?;