-- name: GetUser :one
SELECT * FROM users WHERE user_id = $1;

-- name: CreateUser :one
INSERT INTO users (name, email) VALUES ($1, $2) RETURNING *;

-- name: UpdateUser :one
UPDATE users SET name = $2, email = $3 WHERE user_id = $1 RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users WHERE user_id = $1;

-- name: GetAllUsers :many
SELECT * FROM users;