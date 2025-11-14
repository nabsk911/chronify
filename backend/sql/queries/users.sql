-- name: CreateUser :one
INSERT INTO USERS (email, username, password_hash)
VALUES ($1, $2, $3)
RETURNING id, email, username;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1;
