# 01 · docker-compose con PostgreSQL

- **Fecha:** 2026-10-18
- **Caso de uso / tarea:** 0.3 — docker-compose con PostgreSQL solamente
- **Herramienta y modo:** OpenCode · Plan + Build
- **Modelo:** big-pickle (opencode/big-pickle)
- **MCP usados:** filesystem / web search (versión vigente de PostgreSQL)

## Prompt

> Lee AGENTS.MD, docs/roadmap.md y .env.example (no leas el .env)
>
> La tarea a realizar: el paso 0.3 del roadmap. Crear docker-compose.yml en la raíz con un
> único servicio "postgres" para desarrollo.
>
> Requisitos: imagen oficial de PostgreSQL con versión mayor fija; credenciales por
> POSTGRES_USER / POSTGRES_PASSWORD / POSTGRES_DB del .env; volumen con nombre; healthcheck
> con pg_isready; puerto 5432 solo en 127.0.0.1; restart unless-stopped.
>
> Plan con contenido propuesto, justificación de la versión y comandos de verificación.
> Luego: implementá el plan aprobado.

## Resultado

- `docker-compose.yml` en la raíz: servicio `postgres` con imagen `postgres:18`, healthcheck
  `pg_isready`, volumen nombrado `pgdata` montado en `/var/lib/postgresql` (desde PG 18 el
  `PGDATA` es versionado y el volumen oficial se definió en esa raíz), puerto solo en loopback.
- `docs/roadmap.md`: 0.3 marcado ✅.
- Sin iteraciones de autocorrección: `docker compose config` y `docker compose up` pasaron a la primera.

## Correcciones humanas

- Ninguna.

## Aprendizaje

- Ninguno por ahora.
