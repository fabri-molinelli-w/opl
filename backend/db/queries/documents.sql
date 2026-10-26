-- name: CreateDocument :one
INSERT INTO documents (title, content_md, category, tags, created_by, updated_by)
VALUES (sqlc.arg('title'), sqlc.arg('content_md'), sqlc.arg('category'), sqlc.arg('tags')::text[],
        sqlc.arg('author_id'), sqlc.arg('author_id'))
RETURNING id;

-- name: InsertDocumentVersion :exec
INSERT INTO document_versions (document_id, version, title, content_md, category, tags, edited_by)
VALUES (sqlc.arg('document_id'), sqlc.arg('version'), sqlc.arg('title'), sqlc.arg('content_md'),
        sqlc.arg('category'), sqlc.arg('tags')::text[], sqlc.arg('edited_by'));

-- name: GetDocumentByID :one
SELECT d.id, d.title, d.content_md, d.category, d.tags, d.version,
       d.created_at, d.updated_at, d.created_by,
       cu.name AS created_by_name, uu.name AS updated_by_name
FROM documents d
JOIN users cu ON cu.id = d.created_by
JOIN users uu ON uu.id = d.updated_by
WHERE d.id = $1 AND d.deleted_at IS NULL;

-- name: LockDocumentForUpdate :one
SELECT id, title, content_md, category, tags, version
FROM documents
WHERE id = $1 AND deleted_at IS NULL
FOR UPDATE;

-- name: UpdateDocument :one
UPDATE documents
SET title      = sqlc.arg('title'),
    content_md = sqlc.arg('content_md'),
    category   = sqlc.arg('category'),
    tags       = sqlc.arg('tags')::text[],
    version    = version + 1,
    updated_by = sqlc.arg('updated_by'),
    updated_at = now()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING version;

-- name: ListDocuments :many
-- El orden se elige con CASE sobre valores validados (sin SQL dinámico).
SELECT d.id, d.title, d.category, d.tags, d.version, d.created_at, d.updated_at,
       uu.name AS updated_by_name
FROM documents d
JOIN users uu ON uu.id = d.updated_by
WHERE d.deleted_at IS NULL
  AND (sqlc.narg('q')::text IS NULL
       OR d.title ILIKE '%' || sqlc.narg('q')::text || '%'
       OR d.content_md ILIKE '%' || sqlc.narg('q')::text || '%'
       OR array_to_string(d.tags, ' ') ILIKE '%' || sqlc.narg('q')::text || '%')
  AND (sqlc.narg('category')::text IS NULL OR d.category = sqlc.narg('category')::text)
  AND (sqlc.narg('tag')::text IS NULL OR d.tags @> ARRAY[sqlc.narg('tag')::text]::text[])
ORDER BY
  CASE WHEN sqlc.arg('sort_key')::text = 'title_asc' THEN lower(d.title) END ASC,
  CASE WHEN sqlc.arg('sort_key')::text = 'title_desc' THEN lower(d.title) END DESC,
  CASE WHEN sqlc.arg('sort_key')::text = 'updatedAt_asc' THEN d.updated_at END ASC,
  CASE WHEN sqlc.arg('sort_key')::text = 'updatedAt_desc' THEN d.updated_at END DESC,
  d.id
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountDocuments :one
SELECT count(*)
FROM documents d
WHERE d.deleted_at IS NULL
  AND (sqlc.narg('q')::text IS NULL
       OR d.title ILIKE '%' || sqlc.narg('q')::text || '%'
       OR d.content_md ILIKE '%' || sqlc.narg('q')::text || '%'
       OR array_to_string(d.tags, ' ') ILIKE '%' || sqlc.narg('q')::text || '%')
  AND (sqlc.narg('category')::text IS NULL OR d.category = sqlc.narg('category')::text)
  AND (sqlc.narg('tag')::text IS NULL OR d.tags @> ARRAY[sqlc.narg('tag')::text]::text[]);

-- name: SoftDeleteDocument :execrows
UPDATE documents
SET deleted_at = now(), updated_at = now()
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListDocumentVersions :many
SELECT v.version, v.title, v.created_at, u.name AS edited_by_name
FROM document_versions v
JOIN users u ON u.id = v.edited_by
WHERE v.document_id = $1
ORDER BY v.version DESC;

-- name: GetDocumentVersion :one
SELECT v.version, v.title, v.content_md, v.category, v.tags, v.created_at, u.name AS edited_by_name
FROM document_versions v
JOIN users u ON u.id = v.edited_by
WHERE v.document_id = $1 AND v.version = $2;

-- name: ListDocumentTags :many
SELECT DISTINCT tag::text AS tag
FROM documents d, unnest(d.tags) AS tag
WHERE d.deleted_at IS NULL
ORDER BY tag;

-- name: ListAllDocumentsForIndex :many
-- Usado por OPSBot (CU-09) para reindexar: devuelve todo el contenido vigente.
SELECT id, title, content_md, category, tags, updated_at
FROM documents
WHERE deleted_at IS NULL
ORDER BY updated_at DESC;
