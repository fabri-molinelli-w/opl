-- Consultas agregadas del dashboard (CU-04 RN-1: un único endpoint).

-- name: DashboardServerCounts :many
SELECT environment, location, count(*)::int AS total
FROM servers
WHERE deleted_at IS NULL
GROUP BY environment, location;

-- name: DashboardServiceStatusCounts :many
SELECT s.current_status, count(*)::int AS total
FROM services s
JOIN servers srv ON srv.id = s.server_id AND srv.deleted_at IS NULL
WHERE s.deleted_at IS NULL
GROUP BY s.current_status;

-- name: DashboardProblemServices :many
SELECT s.id, s.name, s.current_status, s.server_id, srv.hostname
FROM services s
JOIN servers srv ON srv.id = s.server_id AND srv.deleted_at IS NULL
WHERE s.deleted_at IS NULL AND s.current_status IN ('DOWN', 'DEGRADED')
ORDER BY CASE s.current_status WHEN 'DOWN' THEN 0 ELSE 1 END, s.name
LIMIT 10;

-- name: DashboardDocumentCategoryCounts :many
SELECT category, count(*)::int AS total
FROM documents
WHERE deleted_at IS NULL
GROUP BY category;

-- name: DashboardDocumentsUpdatedSince :one
SELECT count(*)::int AS total
FROM documents
WHERE deleted_at IS NULL AND updated_at >= $1;

-- name: DashboardRecentDocuments :many
SELECT d.id, d.title, d.updated_at, u.name AS updated_by_name
FROM documents d
JOIN users u ON u.id = d.updated_by
WHERE d.deleted_at IS NULL
ORDER BY d.updated_at DESC
LIMIT 5;

-- name: DashboardActiveIncidentCounts :many
SELECT severity, count(*)::int AS total
FROM incidents
WHERE status IN ('open', 'in_progress')
GROUP BY severity;

-- name: DashboardRecentActiveIncidents :many
SELECT i.id, i.title, i.severity, i.status, i.opened_at, COALESCE(s.name, '')::text AS service_name
FROM incidents i
LEFT JOIN services s ON s.id = i.service_id
WHERE i.status IN ('open', 'in_progress')
ORDER BY i.severity, i.opened_at DESC
LIMIT 5;

-- name: DashboardCheckCounts :one
-- RN-4: uptime % = chequeos UP o DEGRADED / total de chequeos en la ventana.
SELECT count(*)::int AS total,
       count(*) FILTER (WHERE status IN ('UP', 'DEGRADED'))::int AS available
FROM health_checks
WHERE checked_at >= $1;

-- name: DashboardCheckBuckets :many
-- Serie de 24 h en buckets de 15 minutos: disponibilidad y latencia promedio.
SELECT to_timestamp(floor(extract(epoch FROM checked_at) / 900) * 900)::timestamptz AS bucket,
       count(*)::int AS total,
       count(*) FILTER (WHERE status IN ('UP', 'DEGRADED'))::int AS available,
       COALESCE(avg(response_ms) FILTER (WHERE status <> 'DOWN'), 0)::float8 AS avg_response_ms
FROM health_checks
WHERE checked_at >= $1
GROUP BY bucket
ORDER BY bucket;
