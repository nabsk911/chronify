-- name: BulkCreateEvents :copyfrom
INSERT INTO events (
  timeline_id,
  time_marker,
  title,
  subtitle,
  description,
  media_name,
  media_type,
  media_url,
  position
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
);

-- name: GetEventsByTimelineId :many
SELECT *
FROM events
WHERE timeline_id = $1
ORDER BY position ASC;

-- name: GetMaxPositionByTimelineId :one
SELECT COALESCE(MAX(position), 0) 
FROM events
WHERE timeline_id = $1;

-- name: BulkUpdateEvents :batchexec
UPDATE events e
SET 
    time_marker = $3,
    title = $4,
    subtitle = $5,
    description = $6,
    media_name = $7,
    media_type = $8,
    media_url = $9,
    updated_at = CURRENT_TIMESTAMP
FROM timelines t
WHERE e.id = $1
  AND e.timeline_id = t.id
  AND t.user_id = $2;

-- name: DeleteEvent :execrows
DELETE FROM events e
USING timelines t
WHERE e.id = $1
  AND e.timeline_id = t.id
  AND t.user_id = $2;

-- name: DeleteEventsByTimelineId :execrows
DELETE FROM events e
USING timelines t
WHERE e.timeline_id = t.id
  AND e.timeline_id = $1
  AND t.user_id = $2;
