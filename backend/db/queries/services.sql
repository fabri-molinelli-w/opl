-- name: CreateService :one
INSERT INTO services (
    server_id, name, type, port, check_type, check_target,
    expected_status, interval_seconds, latency_threshold_ms
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetServiceByID :one
SELECT *
FROM services
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListServicesByServerID :many
SELECT *
FROM services
WHERE server_id = $1 AND deleted_at IS NULL
ORDER BY name;

-- name: ListServices :many
SELECT s.*
FROM services s
JOIN servers srv ON srv.id = s.server_id AND srv.deleted_at IS NULL
WHERE s.deleted_at IS NULL
  AND (sqlc.narg('server_id')::uuid IS NULL
       OR s.server_id = sqlc.narg('server_id')::uuid)
  AND (sqlc.narg('environment')::text IS NULL
       OR srv.environment = sqlc.narg('environment')::text)
  AND (sqlc.narg('status')::text IS NULL
       OR s.current_status = sqlc.narg('status')::text)
  AND (sqlc.narg('type')::text IS NULL
       OR s.type = sqlc.narg('type')::text)
ORDER BY s.name
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountServices :one
SELECT count(*)
FROM services s
JOIN servers srv ON srv.id = s.server_id AND srv.deleted_at IS NULL
WHERE s.deleted_at IS NULL
  AND (sqlc.narg('server_id')::uuid IS NULL
       OR s.server_id = sqlc.narg('server_id')::uuid)
  AND (sqlc.narg('environment')::text IS NULL
       OR srv.environment = sqlc.narg('environment')::text)
  AND (sqlc.narg('status')::text IS NULL
       OR s.current_status = sqlc.narg('status')::text)
  AND (sqlc.narg('type')::text IS NULL
       OR s.type = sqlc.narg('type')::text);

-- name: UpdateService :one
UPDATE services
SET name                 = COALESCE(sqlc.narg('name'), name),
    type                 = COALESCE(sqlc.narg('type'), type),
    port                 = COALESCE(sqlc.narg('port'), port),
    check_type           = COALESCE(sqlc.narg('check_type'), check_type),
    check_target         = COALESCE(sqlc.narg('check_target'), check_target),
    expected_status      = COALESCE(sqlc.narg('expected_status'), expected_status),
    interval_seconds     = COALESCE(sqlc.narg('interval_seconds'), interval_seconds),
    latency_threshold_ms = COALESCE(sqlc.narg('latency_threshold_ms'), latency_threshold_ms),
    updated_at           = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteService :execrows
UPDATE services
SET deleted_at = now(), updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteServicesByServer :exec
UPDATE services
SET deleted_at = now(), updated_at = now()
WHERE server_id = $1 AND deleted_at IS NULL;
