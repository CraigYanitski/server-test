-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING * ;

-- name: GetUserByEmail :one
SELECT * FROM users 
WHERE email=$1 ;

-- name: ResetUsers :exec
DELETE from users ;

-- name: UpdateUser :one
UPDATE users 
SET updated_at = NOW(),
    email = $2,
    hashed_password = $3
WHERE id = $1
RETURNING * ;

-- name: UpdateUserToRed :one
UPDATE users 
SET is_chirpy_red = true
WHERE id = $1 
RETURNING * ;
