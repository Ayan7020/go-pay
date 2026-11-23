-- name: CreateAccount :one
INSERT INTO accounts (owner_id, name, currency, normal_balance)
VALUES ($1, $2, $3, $4)
RETURNING *;