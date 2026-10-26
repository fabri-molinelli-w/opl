# Roadmap y estado

Actualizar este archivo al cerrar cada tarea. El agente lo lee para saber qué sigue.
Ritmo: **una tarea por sesión de OpenCode**. No se empieza la siguiente hasta que la anterior pasa la verificación.

Estados: ⬜ pendiente · 🟨 en curso · ✅ terminado

## Fase 0 — Contexto y andamiaje

| Estado | Paso | Tarea |
|---|---|---|
| ✅ | 0.1 | AGENTS.md raíz y anidados, opencode.json, .aisignore, specs |
| ✅ | 0.2 | Repo en GitHub, primer commit, `opencode mcp list` con los 4 MCP conectados |
| ✅ | 0.3 | docker-compose con PostgreSQL solamente |
| ✅ | 0.4 | Esqueleto backend: chi, config, slog, conexión a Postgres, goose, `GET /api/health` |
| ✅ | 0.5 | Generación: oapi-codegen + sqlc + Makefile (`make generate`, `make verify`) |
| ✅ | 0.6 | Esqueleto frontend: Vite, Tailwind, Router, TanStack Query, Orval, layout con menú lateral |
| ✅ | 0.7 | CI en GitHub Actions (lint + tests backend y frontend) |

## Fase 1 — Documentación y dashboard (foco actual)

| Estado | CU | Tarea |
|---|---|---|
| ✅ | CU-01 | Login, roles (admin / sysadmin / soporte), gestión de usuarios |
| ✅ | CU-02 | Inventario: primero servidores, después servicios |
| ✅ | CU-05 | Documentación: ABM + búsqueda + categorías y etiquetas |
| ✅ | CU-05 | Documentación: historial de versiones + bloqueo optimista (vínculos con el inventario → Extras) |
| ✅ | CU-04 | Dashboard etapa 1 (inventario + documentación) |
| ✅ | — | Datos semilla ficticios para la demo (inventario, laboratorio y documentos) |

## Fase 2 — Operación

| Estado | CU | Tarea |
|---|---|---|
| ✅ | — | Laboratorio en docker-compose (`lab-nginx`; api-demo y https quedan para CU-06/07) |
| ✅ | CU-03 | Health checks + apertura automática de incidentes |
| ✅ | CU-10 | Gestión de incidentes (sin deploys relacionados ni post-mortem) |
| ✅ | CU-04 | Dashboard etapa 2 (estado de servicios, incidentes, gráficos 24 h) |
| ⬜ | CU-08 | Notificaciones internas (campana) |

## Fase 3 — OPLBot

| Estado | CU | Tarea |
|---|---|---|
| ✅ | CU-09 | ai-service: ingesta de documentos en ChromaDB (sincronización por versión, ADR-023) |
| ✅ | CU-09 | ai-service: chat con RAG + tools de solo lectura, panel lateral en el frontend |

> El chatbot integrado es un criterio de evaluación del TP (hasta 2 puntos). Aunque no sea el foco
> de esta etapa, tiene que estar antes de la entrega final.

## Fase 4 — Completar

| Estado | CU | Tarea |
|---|---|---|
| ⬜ | CU-06 | Certificados SSL/TLS |
| ⬜ | CU-07 | Deploys + webhook |

## Fase 5 — Puesta en producción (Debian)

| Estado | Tarea |
|---|---|
| ⬜ | `docker-compose.prod.yml` sin laboratorio, reverse proxy con HTTPS |
| ⬜ | Backup diario de PostgreSQL y procedimiento de restauración documentado en OPL |
| ⬜ | Guía de instalación en `docs/instalacion-debian.md` |

## Extras

- Migrar OPLBot a LangGraph con paso de verificación.
- Vínculos de documentos con servidores/servicios (`document_links`) y listado en la ficha.
- Búsqueda con `pg_trgm` / `tsvector` (ADR-020).
- Adjuntos en documentos (PDF, imágenes).
- Login con Active Directory / LDAP.
- Email por SMTP.
