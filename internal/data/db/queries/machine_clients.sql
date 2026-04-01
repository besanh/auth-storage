-- name: InsertMachineClient :one
INSERT INTO machine_clients (
    client_id,
    client_secret_hash,
    name,
    scopes,
    created_at,
    updated_at
) VALUES (
    $1,
    $2,
    $3,
    $4,
    now(),
    $5
) RETURNING *;

-- name: GetMachineClientByID :one
SELECT * FROM machine_clients WHERE client_id = $1;

-- name: GetMachineClientByName :one
SELECT * FROM machine_clients WHERE name = $1;

-- name: UpdateMachineClient :one
UPDATE machine_clients
SET
    client_secret_hash = $2,
    name = $3,
    scopes = $4,
    updated_at = now()
WHERE client_id = $1
RETURNING *;

-- name: DeleteMachineClient :one
DELETE FROM machine_clients WHERE client_id = $1 RETURNING *;