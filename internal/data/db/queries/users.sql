-- name: InsertUser :one
INSERT INTO users (
    email,
    password_hash,
    status,
    role,
    scope,
    name,
    avatar_url,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7,
    now(),
    $8
) RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdatePasswordHash :one
UPDATE users
SET
    password_hash = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET
    name = $2,
    avatar_url = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;
