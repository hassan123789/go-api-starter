-- name: GetUserByID :one
SELECT id, email, password_hash, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, password_hash, created_at, updated_at
FROM users
WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (email, password_hash, created_at, updated_at)
VALUES ($1, $2, $3, $4)
RETURNING id, email, password_hash, created_at, updated_at;

-- name: EmailExists :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);
