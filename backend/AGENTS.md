# AGENTS.md — backend (Go)

Aplica a todo lo que esté dentro de `/backend`. Complementa el `AGENTS.md` raíz.

## Estructura

```
backend/
├── cmd/opl/main.go            Punto de entrada: carga config, arma dependencias, levanta HTTP + scheduler
├── internal/
│   ├── config/                Lectura de env vars (struct tipado, falla al iniciar si falta algo)
│   ├── api/
│   │   ├── gen/               GENERADO por oapi-codegen — no editar
│   │   └── handlers/          Implementación de la interfaz StrictServerInterface
│   ├── domain/                Tipos de dominio y errores (ErrNotFound, ErrConflict, ...)
│   ├── service/               Lógica de negocio, un archivo por agregado (server.go, incident.go, ...)
│   ├── repository/            Wrappers finos sobre db/gen si hace falta componer queries
│   ├── db/gen/                GENERADO por sqlc — no editar
│   ├── auth/                  JWT, bcrypt, middleware de autenticación y de rol
│   ├── scheduler/             Pool de workers de health checks + cron de certificados
│   ├── notify/                Notificaciones internas (se guardan en base, el frontend las consulta)
│   └── aiclient/              Cliente HTTP hacia ai-service (/chat, /ingest)
├── db/
│   ├── migrations/            goose: NNNNN_descripcion.sql con -- +goose Up / Down
│   └── queries/               SQL para sqlc, un archivo por tabla/agregado
├── sqlc.yaml
├── oapi-codegen.yaml
└── Makefile
```

## Reglas de Go

- **Errores explícitos**: siempre `if err != nil { return ..., fmt.Errorf("contexto: %w", err) }`. Nunca ignorar un error con `_`.
- **Sin panics** fuera de `main` para errores de arranque.
- **Inyección de dependencias por constructor**: `func NewServerService(repo ServerRepository, log *slog.Logger) *ServerService`. Nada de variables globales ni `init()` con efectos.
- Los services dependen de **interfaces** chicas definidas en el propio paquete del service (para poder testear con fakes).
- `context.Context` es el primer parámetro de toda función que haga I/O.
- Logging solo con `log/slog`, estructurado: `log.Info("health check", "service_id", id, "status", st)`. Nunca loguear contraseñas ni tokens.
- Mapeo de errores de pgx: `pgx.ErrNoRows` → `domain.ErrNotFound`; violación de unique (`23505`) → `domain.ErrConflict`.
- La traducción dominio → HTTP se hace en **un solo lugar** (helper en `handlers/errors.go`).
- Transacciones con `pgx.Tx` cuando una operación escribe en más de una tabla (por ejemplo, cambio de estado de incidente + nota en la línea de tiempo).

## Base de datos

- Tablas en `snake_case` plural (`servers`, `services`, `health_checks`, `incidents`, `incident_events`, `documents`, `document_versions`, `deploys`, `certificates`, `users`).
- Toda tabla tiene `id uuid primary key default gen_random_uuid()`, `created_at timestamptz not null default now()`, `updated_at timestamptz not null default now()`.
- Baja lógica con `deleted_at timestamptz null`; las queries de listado filtran `deleted_at is null`.
- Enums como `text` + `CHECK` (más simple de migrar que `CREATE TYPE`).
- Una migración nunca se edita después de commiteada: se crea una nueva.

## Concurrencia (scheduler)

- Pool de workers con tamaño configurable (`HEALTHCHECK_WORKERS`, default 10) consumiendo de un channel.
- Cada check tiene timeout propio con `context.WithTimeout`.
- El scheduler se detiene limpio con el `context` de `main` (señales SIGINT/SIGTERM).
- Nada de goroutines sueltas sin forma de cancelarlas.

## Tests

- `go test ./...` con tests de tabla (`[]struct{ name string; ... }`).
- Services: tests unitarios con fakes de los repositorios.
- Handlers: `httptest` contra el router.
- Cada caso de uso tiene al menos un test por criterio de aceptación de su spec.

## Verificación (loop de autocorrección)

```bash
make generate
go vet ./...
golangci-lint run
go test ./...
```
