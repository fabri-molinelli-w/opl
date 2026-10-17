# Arquitectura de OPL

## Vista general

```
                    ┌──────────────────────┐
  Navegador ───────▶│  frontend (React)    │  :5173
                    └──────────┬───────────┘
                               │ REST + JWT (cliente generado por Orval)
                    ┌──────────▼───────────┐
  GitHub Actions ──▶│  backend (Go, chi)   │  :8080   ← único punto de entrada
  (webhook deploys) │  handlers → services │
                    │  → repositories      │
                    │  + scheduler         │
                    └───┬──────────────┬───┘
                        │              │ X-Internal-Token + JWT del usuario
              ┌─────────▼──┐   ┌───────▼────────────┐
              │ PostgreSQL │   │ ai-service (OPLBot)│  red interna, sin puerto público
              └────────────┘   │ FastAPI+LangChain  │
                               │ ChromaDB local     │
                               └───────┬────────────┘
                                       │ tools GET /api/... con el JWT del usuario
                                       └──────────▶ backend
                    ┌──────────────────────┐
                    │ lab: nginx, api-demo,│  ← objetivos de los health checks
                    │ https-expiring       │
                    └──────────────────────┘
```

## Componentes

### frontend
SPA en React + TypeScript. Solo habla con el backend. El cliente HTTP y los tipos se generan desde
`api/openapi.yaml` con Orval. OPLBot aparece como panel lateral en todas las pantallas.

### backend (Go)
- **handlers**: implementan la interfaz generada por oapi-codegen (modo strict). Validan y traducen errores.
- **services**: lógica de negocio. Ejemplos: transición de estados de incidentes, cálculo de MTTR, reglas de alertas.
- **repositories**: acceso a PostgreSQL con código generado por sqlc.
- **scheduler**: pool de goroutines que ejecuta health checks HTTP/TCP; cron diario de certificados. Abre incidentes automáticos tras N fallos consecutivos.
- **notify**: genera notificaciones internas (tabla `notifications`) con ventana anti-duplicados; el frontend las consulta.
- **aiclient**: reenvía consultas de chat y avisa cambios de documentos al ai-service.

### ai-service (OPLBot)
- `/ingest`: chunking por encabezados Markdown → embeddings locales → ChromaDB.
- `/chat`: agente LangChain con tools de solo lectura que llaman a endpoints GET del backend **con el JWT del usuario**, por lo que heredan sus permisos.

### lab
Contenedores objetivo para la demo: un Nginx, una API de ejemplo y un endpoint HTTPS con certificado autofirmado
próximo a vencer. Permiten mostrar el ciclo caída → DOWN → incidente → alerta → OPLBot → resolución → post-mortem.

## Seguridad entre servicios
- El frontend nunca habla con ai-service.
- backend → ai-service: header `X-Internal-Token`.
- ai-service → backend: JWT del usuario original (reenviado por el backend).
- ai-service no publica puertos en docker-compose.

## Modelo de datos (resumen)

| Tabla | Campos principales |
|---|---|
| users | id, email (unique), password_hash, name, role (admin/sysadmin/soporte), active |
| servers | id, hostname (unique), ip, environment (prod/staging/dev), os, location (onprem/cloud), owner, tags text[], deleted_at |
| services | id, server_id, name, type (web/db/queue/other), port, check_type (http/tcp), check_target, expected_status, interval_seconds, latency_threshold_ms, current_status, deleted_at |
| health_checks | id, service_id, status (UP/DOWN/DEGRADED), response_ms, error, checked_at |
| incidents | id, title, service_id, severity (SEV1..SEV4), status (open/in_progress/resolved/closed), assignee_id, root_cause, auto_created, opened_at, resolved_at, closed_at |
| incident_events | id, incident_id, author_id, kind (note/status_change/system), body, created_at |
| documents | id, title, content_md, category, tags text[], index_status, created_by, updated_by, deleted_at |
| document_versions | id, document_id, version, title, content_md, edited_by, created_at |
| document_links | document_id, server_id, service_id |
| deploys | id, service_id, version, environment, author, result (success/failed/rollback), notes, deployed_at |
| certificates | id, domain, port, issuer, not_after, days_left, status (ok/warning/critical), last_checked_at |
| notification_prefs | user_id, event_type, enabled |
| notifications | id, user_id, event_type, title, body, link, dedupe_key, read_at, created_at |

## Despliegue en la empresa (Debian)

- Servidor Debian estable (12 o 13) con Docker Engine y el plugin Docker Compose desde el repositorio oficial de Docker.
- Un `docker-compose.prod.yml` sin el laboratorio y con:
  - PostgreSQL con volumen persistente y backup diario (`pg_dump` por cron del host).
  - Backend y frontend detrás de un reverse proxy (Nginx o Caddy) con HTTPS del certificado interno de la empresa.
  - ai-service sin puertos publicados.
- Variables en `/etc/opl/.env` con permisos `600`, fuera del repositorio.
- Logs a stdout (slog JSON) leídos con `docker compose logs` o enviados a journald.
- Solo accesible desde la red interna / VPN.
