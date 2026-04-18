-- name: AddBookmark :exec
INSERT INTO bookmarks (user_id, timeline_id)
VALUES ($1, $2);

-- name: RemoveBookmark :exec
DELETE FROM bookmarks
WHERE user_id = $1 AND timeline_id = $2;

-- name: GetBookmarkedTimelinesByUserId :many
SELECT
    t.id,
    t.user_id,
    t.title,
    t.description,
    t.is_public,
    t.created_at
FROM bookmarks b
JOIN timelines t ON t.id = b.timeline_id
WHERE b.user_id = $1 AND (t.is_public = true OR t.user_id = $1)
ORDER BY t.created_at DESC;

