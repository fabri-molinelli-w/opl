-- name: CreateCertificate :one
INSERT INTO certificates (domain, port) VALUES ($1, $2)
RETURNING id, domain, port, issuer, not_after, days_left, status, error, last_checked_at, created_at;

-- name: GetCertificate :one
SELECT id, domain, port, issuer, not_after, days_left, status, error, last_checked_at, created_at
FROM certificates WHERE id = $1;

-- name: ListCertificates :many
-- Ordenados por días restantes (los que vencen antes, primero).
SELECT id, domain, port, issuer, not_after, days_left, status, error, last_checked_at, created_at
FROM certificates
WHERE sqlc.narg('status')::text IS NULL OR status = sqlc.narg('status')::text
ORDER BY days_left ASC NULLS LAST, domain;

-- name: ListCertificatesDue :many
SELECT id FROM certificates
WHERE last_checked_at IS NULL OR last_checked_at < sqlc.arg('before')::timestamptz;

-- name: UpdateCertificateCheck :exec
UPDATE certificates
SET issuer = sqlc.arg('issuer'), not_after = sqlc.narg('not_after'), days_left = sqlc.narg('days_left'),
    status = sqlc.arg('status'), error = sqlc.narg('error'), last_checked_at = now(), updated_at = now()
WHERE id = sqlc.arg('id');

-- name: DeleteCertificate :execrows
DELETE FROM certificates WHERE id = $1;

-- name: CountCertificatesByStatus :many
SELECT status, count(*)::int AS total FROM certificates GROUP BY status;
