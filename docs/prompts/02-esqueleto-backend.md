# 01 · docker-compose con PostgreSQL

- **Fecha:** 2026-10-18
- **Caso de uso / tarea:** 0.4 — Esqueleto backend: chi, config, slog, conexión a Postgres, goose, `GET /api/health`
- **Herramienta y modo:** OpenCode · Plan + Build
- **Modelo:** big-pickle (opencode/big-pickle)
- **MCP usados:** solo utilizo filesystem

## Prompt

Leé AGENTS.md, backend/AGENTS.md, api/openapi.yaml, docs/roadmap.md y .env.example (no leas .env).

Tarea: paso 0.4 del roadmap. Esqueleto del backend en /backend.

> Alcance:
> - go.mod con módulo github.com/fabri-molinelli-w/opl/backend y la versión de Go instalada (consultá `go version`).
> - cmd/opl/main.go: carga config, arma dependencias, levanta el servidor HTTP y se apaga limpio con SIGINT/SIGTERM.
> - internal/config: lee HTTP_PORT y DATABASE_URL de variables de entorno; si falta alguna, falla al arrancar con un mensaje claro.
> - Logging con log/slog en formato JSON.
> - Conexión a PostgreSQL con pgxpool (pgx v5).
> - goose integrado: las migraciones de backend/db/migrations se embeben en el binario y se aplican al arrancar. Por ahora no hay > migraciones.
> - Router chi con GET /api/health que responde según el schema Health de api/openapi.yaml (status y database, con ping a la base).
> - Un test con httptest para /api/health.

> Fuera de alcance: oapi-codegen, sqlc, Makefile, Dockerfile, autenticación. En este paso el handler de
> /api/health se escribe a mano; en el paso 0.5 se reemplaza por código generado.

> Antes de proponer dependencias, consultá con context7 la API actual de chi, pgx v5 y goose.

> No escribas código. Entregá un plan con: árbol de archivos, dependencias con versión,
> responsabilidad de cada archivo y comandos de verificación.

## Resultado

- go vet y go test pasan en backend/. El esqueleto del backend está implementado conforme al plan (config, slog JSON, pgxpool, goose con migraciones embebidas, health handler manual, shutdown limpio, tests httptest).
- error por una limitación de Go: no acepta rutas con '..'. Únicamente puede embeber archivos que estén dentro de la carpeta
del propio paquete o en subcarpetas.
    - La solución fue
    1) Crear backend/db/embed.go
    2) Creá backend/db/migrations/00001_baseline.sql como migración inicial vacía
    3) En cmd/opl/main.go sacá el //go:embed y usá db.Migrations para goose.
        Si goose necesita *sql.DB, obtenelo del pool con stdlib.OpenDBFromPool de pgx/v5 (consultá context7).
    4) Corré go vet ./... y go test ./... en backend/ hasta que pasen.

## Correcciones humanas

- Ninguna.

## Aprendizaje

- Las limitaciones de Go respecto a las rutas que manipula
