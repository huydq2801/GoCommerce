-- name: CreateUser :one
INSERT INTO users (user_email,user_password,user_fullname,user_age,user_status,user_level)
VALUES ($1,$2,$3,$4,$5,$6) RETURNING *;

-- name: ListUsers :many
SELECT * FROM users WHERE user_deleted_at IS NULL
AND (sqlc.arg(search)::text = '' OR user_email ILIKE '%' || sqlc.arg(search) || '%' OR user_fullname ILIKE '%' || sqlc.arg(search) || '%')
ORDER BY user_id DESC LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: CountUsers :one
SELECT count(*) FROM users WHERE user_deleted_at IS NULL
AND (sqlc.arg(search)::text = '' OR user_email ILIKE '%' || sqlc.arg(search) || '%' OR user_fullname ILIKE '%' || sqlc.arg(search) || '%');

-- name: GetUser :one
SELECT * FROM users WHERE uuid=$1 AND user_deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE users SET user_fullname=coalesce(sqlc.narg(user_fullname)::text,user_fullname),
 user_email=coalesce(sqlc.narg(user_email)::text,user_email),
 user_age=coalesce(sqlc.narg(user_age)::int,user_age),
 user_password=coalesce(sqlc.narg(user_password)::text,user_password),
 user_status=coalesce(sqlc.narg(user_status)::int,user_status),
 user_level=coalesce(sqlc.narg(user_level)::int,user_level)
WHERE uuid=sqlc.arg(uuid)::uuid AND user_deleted_at IS NULL RETURNING *;

-- name: SoftDeleteUser :execrows
UPDATE users SET user_deleted_at=now() WHERE uuid=$1 AND user_deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE user_email=$1 AND user_deleted_at IS NULL;
