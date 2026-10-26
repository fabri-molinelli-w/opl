# CU-02 · Inventario de servidores y servicios

## Resumen

Implementa el ABM y la consulta de servidores y de los servicios que corren en ellos, con búsqueda,
filtros combinables, paginación, baja lógica en cascada y datos semilla para la demo. Incluye backend Go,
frontend React y el contrato OpenAPI; `ai-service` no participa en este caso de uso.

Roles: `soporte` puede consultar; `sysadmin` (y `admin` por jerarquía) puede crear, editar y dar de baja.
La autorización se declara por operación con `x-min-role` y la valida el middleware central.

## Qué cambió y por qué

### Contrato → código generado
- `api/openapi.yaml` agrega los tags `servers` y `services` con 10 operaciones y sus schemas
  (`Server`, `Service`, requests de alta/edición, paginados y enums).
- Se mantiene el flujo spec-first: `make generate` regenera oapi-codegen (servidor Go strict),
  sqlc y Orval (cliente TypeScript). Los nombres públicos son camelCase (`pageSize`, `serverId`,
  `checkType`, `currentStatus`); las columnas y queries SQL siguen en snake_case.
- El test `policy_contract_test.go` compara cada operación del contrato con `auth.Policy`, de modo que
  un endpoint nuevo no puede quedar sin rol declarado.

### Migraciones y queries
- `00003_create_servers.sql`: tabla `servers` con `ip inet`, `tags text[]`, checks de `environment` y
  `location`, baja lógica y índice único parcial `lower(hostname) WHERE deleted_at IS NULL`.
  Índice GIN para tags e índices para los filtros.
- `00004_create_services.sql`: tabla `services` con FK a `servers`, checks de `type`, `port`,
  `check_type` y `current_status`, defaults (200/30/1000/UNKNOWN), baja lógica e índices.
- `db/queries/servers.sql` y `services.sql`: filtros opcionales con `sqlc.narg`, paginación,
  lecturas con `host(ip)` y escrituras con `::inet`, y `deleted_at IS NULL` en todos los listados.

### Backend (Go)
- `internal/domain`: tipos de inventario, normalización/validación de enums y `FieldError`/`Details`
  dentro de `domain.Error`.
- `internal/repository`: wrappers de sqlc; traduce `pgx.ErrNoRows` a 404 y `23505` a 409. La baja de un
  servidor y sus servicios ocurre en una sola transacción `pgx.Tx` (RN-6).
- `internal/service`: reglas de negocio (IP IPv4/IPv6 con `netip`, hostname, defaults de servicio,
  targets HTTP válidos, paginación 20/100, merge de PATCH y conflictos).
- `internal/api/handlers`: handlers generados por oapi-codegen; `Deps` inyecta los services.
  El mapeo dominio→HTTP con `details[].field` está centralizado en `handlers/errors.go`.
- `internal/auth`: 10 operaciones nuevas en la política; lectura `soporte`, escritura `sysadmin`.

### Frontend (React + TypeScript)
- `InventoryPage`: tabla de servidores, buscador y filtros combinables (`q`, `environment`, `os`, `tag`)
  y paginación reflejados en la URL; acciones de escritura según rol.
- `ServerDetailPage`: ficha del servidor con sus servicios y CRUD de servicios.
- `ServerForm` y `ServiceForm`: alta/edición con errores por campo provenientes de `details`.
- Solo se usan hooks y tipos generados por Orval; sin `any`.

### Datos semilla
- `backend/cmd/seed` + `make seed`: 5 servidores ficticios con IPs de documentación y un servicio cada uno.
  Es manual, idempotente (UPSERT por hostname y servicios solo si faltan) y solo para desarrollo;
  no se ejecuta desde migraciones ni al arrancar el backend.

## Cómo probarlo

Requisitos: Docker Compose con PostgreSQL y las migraciones aplicadas.

```bash
docker compose up -d postgres
make generate
make seed

# Backend
cd backend && go vet ./... && golangci-lint run && go test ./...

# Integración con Postgres real (crea opl_test)
make test-integration

# Frontend
cd frontend && npm run lint && npm run typecheck && npm run test -- --run

# Verificación completa
make verify
make check-generated
```

Prueba manual:
1. Iniciar backend y frontend (`make run-backend`, `make run-frontend`) e ingresar como `soporte`:
   ver el listado, combinar filtros y abrir la ficha de un servidor.
2. Ingresar como `sysadmin`: crear un servidor, editar un servicio y dar de baja un servidor.
3. Verificar que el servidor dado de baja y sus servicios desaparecen de los listados.
4. Repetir `make seed` y confirmar que no se duplican servidores ni servicios.

## Criterios de aceptación cubiertos

- **CA-1 · Los filtros se combinan** (`environment=prod&tag=web`):
  `backend/internal/repository/inventory_integration_test.go` → `TestInventoryFiltersPaginationAndSoftDelete/filters combine`;
  `frontend/src/features/inventory/InventoryPage.test.tsx` → `actualiza los filtros en la URL y reinicia la página`.
- **CA-2 · IP inválida → 400 con `details[].field = "ip"`**:
  `backend/internal/service/inventory_test.go` → `TestCreateServer_InvalidIP`;
  `backend/internal/api/handlers/inventory_test.go` → `TestCreateServerHandler_ValidationDetails`.
- **CA-3 · Hostname repetido → 409**:
  `backend/internal/service/inventory_test.go` → `TestCreateServer_DuplicateHostnameConflict`;
  `backend/internal/api/handlers/inventory_test.go` → `TestCreateServerHandler_DuplicateHostname`;
  `backend/internal/repository/inventory_integration_test.go` → `TestInventoryFiltersPaginationAndSoftDelete/hostname uniqueness is case-insensitive`.
- **CA-4 · Un servidor dado de baja no aparece en listados**:
  `backend/internal/repository/inventory_integration_test.go` → `TestInventoryFiltersPaginationAndSoftDelete/soft deleting a server hides it and its services`
  (verifica `Get` → 404, listado de servidores sin el eliminado y servicios del servidor ocultos).
  Complementado por `TestDeleteServiceIsSoft`.

También se cubren las reglas de negocio con tests de servicio para default de `location`, paginación,
defaults de servicio (`expectedStatus`, `intervalSeconds`, `latencyThresholdMs`, `UNKNOWN`), target HTTP
inválido, servidor inexistente y puerto fuera de rango.

## Verificación ejecutada

- `make verify`: backend `go vet`, `golangci-lint` (0 issues) y `go test ./...`; frontend `eslint`,
  `tsc --noEmit` y Vitest (8 archivos, 19 tests) en verde.
- `make test-integration`: en verde contra el PostgreSQL de Docker Compose.
- `make check-generated`: sin diferencias después de regenerar contrato, sqlc y Orval.
- `make seed`: ejecutado dos veces sin duplicados.
