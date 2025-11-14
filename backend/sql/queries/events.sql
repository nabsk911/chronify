-- name: BulkCreateEvents :copyfrom
INSERT INTO events ( timeline_id, title, card_title, card_subtitle, card_detailed_text)
VALUES ( $1, $2, $3, $4, $5);

-- name: GetEventsByTimelineId :many
SELECT * FROM events
WHERE timeline_id = $1
ORDER BY created_at ASC;

-- name: BulkUpdateEvents :batchexec
UPDATE events
SET 
    title = $2, 
    card_title = $3, 
    card_subtitle = $4, 
    card_detailed_text = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: DeleteEvent :exec
DELETE FROM events
WHERE id = $1;


