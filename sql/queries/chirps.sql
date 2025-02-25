-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING * ;

-- name: ResetChirps :exec
DELETE from chirps ;

-- name: GetChirps :many
SELECT * FROM chirps 
ORDER BY created_at ;

-- name: GetChirpsByUser :many
SELECT * FROM chirps 
WHERE user_id = $1 ;

-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1 ;

-- name: DeleteChirp :one
DELETE FROM chirps
WHERE id = $1
RETURNING * ;
