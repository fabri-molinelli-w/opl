-- name: CreateIncident :one
INSERT INTO incidents (title, description, service_id, severity, assignee_id)
VALUES (sqlc.arg('title'), sqlc.arg('description'), sqlc.narg('service_id'), sqlc.arg('severity'),
        sqlc.narg('assignee_id'))
RETURNING id;

-- name: CreateAutoIncident :one
-- RN-4: el índice único parcial garantiza un solo incidente automático abierto por servicio.
INSERT INTO incidents (title, description, service_id, severity, auto_created)
VALUES (sqlc.arg('title'), sqlc.arg('description'), sqlc.arg('service_id'), 'SEV2', true)
ON CONFLICT (service_id) WHERE auto_created AND status <> 'closed' DO NOTHING
RETURNING id;

-- name: GetOpenAutoIncidentForService :one
SELECT id, status
FROM incidents
WHERE service_id = $1 AND auto_created AND status <> 'closed';

-- name: InsertIncidentEvent :exec
INSERT INTO incident_events (incident_id, author_id, kind, body)
VALUES (sqlc.arg('incident_id'), sqlc.narg('author_id'), sqlc.arg('kind'), sqlc.arg('body'));

-- name: GetIncident :one
SELECT i.id, i.title, i.description, i.service_id, i.severity, i.status, i.assignee_id, i.root_cause,
       i.auto_created, i.opened_at, i.resolved_at, i.closed_at, i.created_at, i.updated_at,
       COALESCE(s.name, '')::text AS service_name,
       COALESCE(u.name, '')::text AS assignee_name
FROM incidents i
LEFT JOIN services s ON s.id = i.service_id
LEFT JOIN users u ON u.id = i.assignee_id
WHERE i.id = $1;

-- name: LockIncidentStatus :one
SELECT status FROM incidents WHERE id = $1 FOR UPDATE;

-- name: UpdateIncidentStatus :exec
UPDATE incidents
SET status      = sqlc.arg('status'),
    root_cause  = COALESCE(sqlc.narg('root_cause'), root_cause),
    resolved_at = CASE WHEN sqlc.arg('status')::text = 'resolved' THEN now()
                       WHEN sqlc.arg('status')::text = 'in_progress' THEN NULL
                       ELSE resolved_at END,
    closed_at   = CASE WHEN sqlc.arg('status')::text = 'closed' THEN now() ELSE closed_at END,
    updated_at  = now()
WHERE id = sqlc.arg('id');

-- name: UpdateIncident :execrows
UPDATE incidents
SET title       = COALESCE(sqlc.narg('title'), title),
    description = COALESCE(sqlc.narg('description'), description),
    severity    = COALESCE(sqlc.narg('severity'), severity),
    assignee_id = COALESCE(sqlc.narg('assignee_id'), assignee_id),
    updated_at  = now()
WHERE id = sqlc.arg('id');

-- name: ListIncidents :many
SELECT i.id, i.title, i.service_id, i.severity, i.status, i.auto_created, i.opened_at, i.resolved_at,
       COALESCE(s.name, '')::text AS service_name,
       COALESCE(u.name, '')::text AS assignee_name
FROM incidents i
LEFT JOIN services s ON s.id = i.service_id
LEFT JOIN users u ON u.id = i.assignee_id
WHERE (sqlc.narg('status')::text IS NULL OR i.status = sqlc.narg('status')::text)
  AND (sqlc.narg('severity')::text IS NULL OR i.severity = sqlc.narg('severity')::text)
  AND (sqlc.narg('service_id')::uuid IS NULL OR i.service_id = sqlc.narg('service_id')::uuid)
  AND (NOT sqlc.arg('only_active')::boolean OR i.status IN ('open', 'in_progress'))
ORDER BY i.opened_at DESC, i.id
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountIncidents :one
SELECT count(*)
FROM incidents i
WHERE (sqlc.narg('status')::text IS NULL OR i.status = sqlc.narg('status')::text)
  AND (sqlc.narg('severity')::text IS NULL OR i.severity = sqlc.narg('severity')::text)
  AND (sqlc.narg('service_id')::uuid IS NULL OR i.service_id = sqlc.narg('service_id')::uuid)
  AND (NOT sqlc.arg('only_active')::boolean OR i.status IN ('open', 'in_progress'));

-- name: ListIncidentEvents :many
SELECT e.id, e.kind, e.body, e.created_at, COALESCE(u.name, 'Sistema')::text AS author_name
FROM incident_events e
LEFT JOIN users u ON u.id = e.author_id
WHERE e.incident_id = $1
ORDER BY e.created_at, e.id;
