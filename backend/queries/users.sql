-- name: CreateUser :one
INSERT INTO users (
  username, email, image_url, clerk_id
) VALUES (
  $1, $2, $3, $4
)
ON CONFLICT (clerk_id) DO UPDATE SET
  username = EXCLUDED.username,
  email = EXCLUDED.email,
  image_url = EXCLUDED.image_url
RETURNING *;


-- name: GetUser :one
SELECT * FROM users 
WHERE clerk_id = $1 LIMIT 1;
