-- name: NotifyUsersByRoles :execrows
INSERT INTO notifications (user_id, event_type, title, body, link, dedupe_key)
SELECT u.id, sqlc.arg('event_type'), sqlc.arg('title'), sqlc.arg('body'), sqlc.arg('link'), sqlc.arg('dedupe_key')
FROM users u
WHERE u.active AND u.role = ANY(sqlc.arg('roles')::text[])
ON CONFLICT (user_id, dedupe_key) WHERE read_at IS NULL DO NOTHING;

-- name: NotifyUser :execrows
INSERT INTO notifications (user_id, event_type, title, body, link, dedupe_key)
SELECT u.id, sqlc.arg('event_type'), sqlc.arg('title'), sqlc.arg('body'), sqlc.arg('link'), sqlc.arg('dedupe_key')
FROM users u
WHERE u.id = sqlc.arg('user_id') AND u.active
ON CONFLICT (user_id, dedupe_key) WHERE read_at IS NULL DO NOTHING;

-- name: ListNotifications :many
SELECT id, event_type, title, body, link, read_at, created_at
FROM notifications
WHERE user_id = sqlc.arg('user_id')
  AND (NOT sqlc.arg('unread_only')::boolean OR read_at IS NULL)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit');

-- name: CountUnreadNotifications :one
SELECT count(*)::int FROM notifications WHERE user_id = $1 AND read_at IS NULL;

-- name: MarkNotificationRead :execrows
UPDATE notifications SET read_at = now()
WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id') AND read_at IS NULL;

-- name: NotificationExists :one
SELECT EXISTS (SELECT 1 FROM notifications WHERE id = sqlc.arg('id') AND user_id = sqlc.arg('user_id'));

-- name: MarkAllNotificationsRead :execrows
UPDATE notifications SET read_at = now() WHERE user_id = $1 AND read_at IS NULL;

-- name: DeleteNotificationsBefore :execrows
DELETE FROM notifications WHERE created_at < $1;
