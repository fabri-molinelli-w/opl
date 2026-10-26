-- name: ListDueServices :many
-- Servicios activos cuyo próximo chequeo ya venció (RN-4 de CU-03).
SELECT s.id
FROM services s
JOIN servers srv ON srv.id = s.server_id AND srv.deleted_at IS NULL
WHERE s.deleted_at IS NULL
  AND s.check_enabled
  AND (s.last_checked_at IS NULL
       OR s.last_checked_at + make_interval(secs => s.interval_seconds) <= now())
ORDER BY s.last_checked_at NULLS FIRST
LIMIT sqlc.arg('limit');

-- name: GetServiceCheckConfig :one
SELECT id, name, check_type, check_target, expected_status, latency_threshold_ms
FROM services
WHERE id = $1 AND deleted_at IS NULL;

-- name: LockServiceCheckState :one
SELECT id, server_id, name, current_status, consecutive_failures
FROM services
WHERE id = $1 AND deleted_at IS NULL
FOR UPDATE;

-- name: UpdateServiceCheckState :one
UPDATE services
SET current_status       = sqlc.arg('status'),
    last_checked_at      = now(),
    consecutive_failures = CASE WHEN sqlc.arg('status')::text = 'DOWN'
                                THEN consecutive_failures + 1 ELSE 0 END,
    updated_at           = now()
WHERE id = sqlc.arg('id')
RETURNING consecutive_failures;

-- name: InsertHealthCheck :one
INSERT INTO health_checks (service_id, status, response_ms, http_status, error)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, service_id, status, response_ms, http_status, error, checked_at;

-- name: ListHealthChecks :many
SELECT id, service_id, status, response_ms, http_status, error, checked_at
FROM health_checks
WHERE service_id = $1
ORDER BY checked_at DESC
LIMIT sqlc.arg('limit');

-- name: DeleteHealthChecksBefore :execrows
DELETE FROM health_checks WHERE checked_at < $1;
