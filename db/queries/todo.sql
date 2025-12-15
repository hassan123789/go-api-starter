-- name: GetTodoByID :one
SELECT id, user_id, title, completed, created_at, updated_at
FROM todos
WHERE id = $1;

-- name: GetTodosByUserID :many
SELECT id, user_id, title, completed, created_at, updated_at
FROM todos
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CreateTodo :one
INSERT INTO todos (user_id, title, completed, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, title, completed, created_at, updated_at;

-- name: UpdateTodo :one
UPDATE todos
SET title = $2, completed = $3, updated_at = $4
WHERE id = $1
RETURNING id, user_id, title, completed, created_at, updated_at;

-- name: DeleteTodo :exec
DELETE FROM todos WHERE id = $1;
