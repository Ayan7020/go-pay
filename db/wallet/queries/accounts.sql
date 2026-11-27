-- name: CreateAccount :one
INSERT INTO accounts (kind,owner_id, name, currency, normal_balance)
VALUES ($1,$2, $3, $4, $5)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts 
WHERE id = $1;

-- name: GetSystemAccount :one
SELECT * FROM accounts
WHERE kind = 'system' and name = $1 and currency = $2;

-- name: ListAccountsByOwner :many
SELECT * FROM accounts 
WHERE owner_id = @owner_id::uuid
ORDER BY created_at;

-- name: GetAccountBalance :one
select * from account_balances
where account_id = $1;