-- name: CreateTimeline :one
INSERT INTO TIMELINES (user_id, title, description,is_public)
VALUES ($1, $2, $3, $4)
RETURNING id, user_id, title, description, is_public, created_at;

-- name: GetTimeLineById :one
SELECT * FROM timelines
WHERE id = $1;

-- name: GetTimelinesByUserId :many
SELECT * FROM timelines
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetTimelinesByUserIdAndTitle :many
SELECT * FROM timelines
WHERE user_id = $1 AND title ILIKE '%' || $2 || '%'
ORDER BY created_at DESC;

-- name: GetPublicTimelines :many
SELECT *
FROM timelines
WHERE is_public = true
  AND user_id != $1
ORDER BY created_at DESC;

-- name: GetPublicTimelinesByTitle :many
SELECT * FROM timelines
WHERE is_public = true AND user_id != $2 AND title ILIKE '%' || $1 || '%'
ORDER BY created_at DESC;

-- name: UpdateTimeline :one
UPDATE timelines
SET title = $3,
    description = $4,
    is_public = $5
WHERE id = $1
  AND user_id = $2
RETURNING *;

-- name: DeleteTimeline :execrows
DELETE FROM timelines
WHERE id = $1
  AND user_id = $2;

