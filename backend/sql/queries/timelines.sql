-- name: CreateTimeline :one
INSERT INTO TIMELINES (user_id, title, description)
VALUES ($1, $2, $3)
RETURNING id, user_id, title, description, created_at;

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

-- name: UpdateTimeline :one
UPDATE timelines
SET title = $2, description = $3
WHERE id = $1
RETURNING id, user_id, title, description;

-- name: DeleteTimeline :exec
DELETE FROM timelines
WHERE id = $1;

