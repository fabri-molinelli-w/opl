-- name: CreateServer :one
INSERT INTO servers (hostname, ip, environment, os, location, owner, tags)
VALUES ($1, $2::inet, $3, $4, $5, $6, $7)
RETURNING id, hostname, host(ip) AS ip, environment, os, location, owner, tags, created_at, updated_at;

-- name: GetServerByID :one
SELECT id, hostname, host(ip) AS ip, environment, os, location, owner, tags, created_at, updated_at
FROM servers
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListServers :many
SELECT id, hostname, host(ip) AS ip, environment, os, location, owner, tags, created_at, updated_at
FROM servers
WHERE deleted_at IS NULL
  AND (sqlc.narg('q')::text IS NULL
       OR hostname ILIKE '%' || sqlc.narg('q')::text || '%'
       OR owner ILIKE '%' || sqlc.narg('q')::text || '%'
       OR host(ip) ILIKE '%' || sqlc.narg('q')::text || '%')
  AND (sqlc.narg('environment')::text IS NULL
       OR environment = sqlc.narg('environment')::text)
  AND (sqlc.narg('os')::text IS NULL OR os = sqlc.narg('os')::text)
  AND (sqlc.narg('tag')::text IS NULL
       OR tags @> ARRAY[sqlc.narg('tag')::text]::text[])
ORDER BY hostname
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountServers :one
SELECT count(*)
FROM servers
WHERE deleted_at IS NULL
  AND (sqlc.narg('q')::text IS NULL
       OR hostname ILIKE '%' || sqlc.narg('q')::text || '%'
       OR owner ILIKE '%' || sqlc.narg('q')::text || '%'
       OR host(ip) ILIKE '%' || sqlc.narg('q')::text || '%')
  AND (sqlc.narg('environment')::text IS NULL
       OR environment = sqlc.narg('environment')::text)
  AND (sqlc.narg('os')::text IS NULL OR os = sqlc.narg('os')::text)
  AND (sqlc.narg('tag')::text IS NULL
       OR tags @> ARRAY[sqlc.narg('tag')::text]::text[]);

-- name: UpdateServer :one
UPDATE servers
SET hostname    = COALESCE(sqlc.narg('hostname'), hostname),
    ip          = COALESCE(sqlc.narg('ip')::inet, ip),
    environment = COALESCE(sqlc.narg('environment'), environment),
    os          = COALESCE(sqlc.narg('os'), os),
    location    = COALESCE(sqlc.narg('location'), location),
    owner       = COALESCE(sqlc.narg('owner'), owner),
    tags        = COALESCE(sqlc.narg('tags'), tags),
    updated_at  = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, hostname, host(ip) AS ip, environment, os, location, owner, tags, created_at, updated_at;

-- name: SoftDeleteServer :execrows
UPDATE servers
SET deleted_at = now(), updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;
