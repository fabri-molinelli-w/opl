-- name: CreateDeploy :one
INSERT INTO deploys (service_id, version, environment, author, result, notes, source, deployed_at)
VALUES (sqlc.arg('service_id'), sqlc.arg('version'), sqlc.arg('environment'), sqlc.arg('author'),
        sqlc.arg('result'), sqlc.arg('notes'), sqlc.arg('source'),
        COALESCE(sqlc.narg('deployed_at')::timestamptz, now()))
RETURNING id;

-- name: ListDeploys :many
SELECT d.id, d.service_id, s.name AS service_name, d.version, d.environment, d.author, d.result,
       d.notes, d.source, d.deployed_at
FROM deploys d
JOIN services s ON s.id = d.service_id
WHERE (sqlc.narg('service_id')::uuid IS NULL OR d.service_id = sqlc.narg('service_id')::uuid)
  AND (sqlc.narg('environment')::text IS NULL OR d.environment = sqlc.narg('environment')::text)
  AND (sqlc.narg('from_ts')::timestamptz IS NULL OR d.deployed_at >= sqlc.narg('from_ts')::timestamptz)
  AND (sqlc.narg('to_ts')::timestamptz IS NULL OR d.deployed_at <= sqlc.narg('to_ts')::timestamptz)
ORDER BY d.deployed_at DESC
LIMIT 100;

-- name: FindServiceForDeploy :one
-- RN-3 (CU-07): el webhook identifica el servicio por nombre + environment de su servidor.
SELECT s.id
FROM services s
JOIN servers srv ON srv.id = s.server_id AND srv.deleted_at IS NULL
WHERE s.deleted_at IS NULL AND lower(s.name) = lower(sqlc.arg('name')) AND srv.environment = sqlc.arg('environment')
ORDER BY s.created_at
LIMIT 1;

-- name: ServiceEnvironment :one
SELECT srv.environment
FROM services s JOIN servers srv ON srv.id = s.server_id
WHERE s.id = $1 AND s.deleted_at IS NULL;
