---
name: migracion-sqlc
description: Cómo crear una migración goose y las queries sqlc asociadas en el backend de OPL, respetando las convenciones de tablas. Usar al agregar o cambiar tablas, columnas o consultas SQL.
compatibility: opencode
metadata:
  opencode/slash: "true"
---

## Cuándo usarme
Al crear o modificar tablas, columnas, índices o queries SQL del backend.

## Pasos

1. Listá `backend/db/migrations/` y tomá el siguiente número.
2. Creá `backend/db/migrations/NNNNN_descripcion_en_ingles.sql`:

   ```sql
   -- +goose Up
   CREATE TABLE servers (
       id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
       hostname    text NOT NULL,
       environment text NOT NULL CHECK (environment IN ('prod','staging','dev')),
       created_at  timestamptz NOT NULL DEFAULT now(),
       updated_at  timestamptz NOT NULL DEFAULT now(),
       deleted_at  timestamptz
   );
   CREATE UNIQUE INDEX servers_hostname_uq ON servers (hostname) WHERE deleted_at IS NULL;

   -- +goose Down
   DROP TABLE servers;
   ```

3. Nunca edites una migración ya commiteada; creá una nueva.
4. Queries en `backend/db/queries/<tabla>.sql` con anotaciones sqlc:

   ```sql
   -- name: GetServer :one
   SELECT * FROM servers WHERE id = $1 AND deleted_at IS NULL;

   -- name: ListServers :many
   SELECT * FROM servers
   WHERE deleted_at IS NULL
     AND (sqlc.narg('environment')::text IS NULL OR environment = sqlc.narg('environment'))
   ORDER BY hostname
   LIMIT $1 OFFSET $2;
   ```

5. `make generate` y revisá los tipos generados en `backend/internal/db/gen/`.
6. Corré `go test ./...` para confirmar que los repositorios compilan con los tipos nuevos.

## Convenciones
- Tablas en plural snake_case; enums como `text` + `CHECK`.
- `id`, `created_at`, `updated_at` en todas; `deleted_at` en entidades con baja lógica.
- Índice en toda FK y en columnas de filtro frecuente.
- Filtros opcionales con `sqlc.narg`.
