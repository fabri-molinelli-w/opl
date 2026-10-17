# AGENTS.md — OPL (Operations Pilot Library)

> Fuente única de verdad para cualquier agente que trabaje en este repositorio.
> Leé este archivo completo antes de proponer o escribir cambios.

## 1. Qué es el proyecto

OPL es un portal web de operaciones de infraestructura orientado a equipos de **Soporte IT y SysAdmin**:
base de documentación (procedimientos, runbooks, guías), inventario de servidores y servicios, dashboard,
health checks, incidentes, certificados SSL/TLS, registro de deploys y notificaciones internas.

**Foco de la etapa actual:** documentación (CU-05) y dashboard (CU-04).

**Destino de producción:** servidor Debian con Docker Compose, en la red interna de la empresa. Incluye **OPLBot**, un asistente
(LangChain + RAG sobre la documentación) que **solo consulta** datos, nunca los modifica.

- Casos de uso: `docs/specs/CU-XX-*.md` (uno por archivo).
- Arquitectura: `docs/arquitectura.md`.
- Orden de trabajo y estado: `docs/roadmap.md`.
- Decisiones tomadas: `docs/decisiones.md`.

## 2. Estructura del monorepo

```
/api/openapi.yaml      Contrato OpenAPI 3 — FUENTE DE VERDAD de la API
/backend               Go: API REST, auth, scheduler de health checks, notificaciones
/frontend              React + TypeScript + Vite (SPA)
/ai-service            Python: FastAPI + LangChain + ChromaDB (OPLBot) — red interna
/lab                   Servicios de laboratorio para la demo (nginx, api demo, https)
/docs/specs            Especificaciones por caso de uso
/docs/prompts          Bitácora de prompts relevantes
/.opencode             Comandos y skills de OpenCode
docker-compose.yml     Levanta todo con un comando
```

Cada servicio tiene su propio `AGENTS.md` con las convenciones del lenguaje. Se cargan vía
`instructions` en `opencode.json`; respetalos siempre que toques ese directorio.

## 3. Stack (no cambiar sin aprobación explícita)

| Capa | Tecnología |
|---|---|
| Backend | Go, chi, pgx v5 + sqlc, goose (migraciones), log/slog, golang-jwt v5, bcrypt, robfig/cron v3 |
| Contrato | OpenAPI 3.0 spec-first → oapi-codegen (servidor Go, modo strict) y Orval (cliente TS) |
| Frontend | React + TypeScript strict + Vite, Tailwind CSS, Recharts, TanStack Query, React Router |
| IA | Python 3.12, FastAPI, LangChain, ChromaDB, sentence-transformers (multilingual-e5-small) |
| LLM OPLBot | Modelo gratuito de OpenCode Zen vía `ChatOpenAI` (endpoint compatible OpenAI), por env vars |
| Base de datos | PostgreSQL |
| Infra | Docker + Docker Compose, GitHub Actions |

Las versiones exactas son las de `backend/go.mod`, `frontend/package.json` y `ai-service/pyproject.toml`.
**Nunca** asumas una versión: leé esos archivos.

## 4. Arquitectura en capas (backend)

```
HTTP request → handler (generado por oapi-codegen) → service (lógica de negocio) → repository (sqlc) → PostgreSQL
```

- **Handlers**: validan entrada, llaman al service, traducen errores de dominio a HTTP. Sin SQL, sin lógica de negocio.
- **Services**: reglas de negocio, transacciones, permisos finos. Reciben dependencias por constructor.
- **Repositories**: solo acceso a datos, código generado por sqlc a partir de `backend/db/queries/*.sql`.
- **Scheduler** (health checks, certificados) corre en el mismo binario y reutiliza la capa de services.
- El frontend habla **solo** con el backend Go. El backend es el único que habla con `ai-service`.

## 5. Mapeo de errores a HTTP (obligatorio en todo el backend)

| Situación | Código | Error de dominio Go |
|---|---|---|
| Payload inválido / validación | 400 | `ErrValidation` |
| Sin token o token inválido | 401 | `ErrUnauthenticated` |
| Rol sin permiso | 403 | `ErrForbidden` |
| Recurso inexistente | 404 | `ErrNotFound` |
| Conflicto (duplicado, transición de estado inválida) | 409 | `ErrConflict` |
| Error inesperado | 500 | cualquier otro (se loguea, no se expone el detalle) |

Todas las respuestas de error usan el schema `Error` de `api/openapi.yaml`:
`{ "code": "string", "message": "string", "details": [ { "field": "...", "issue": "..." } ] }`.

## 6. Roles

`admin` > `sysadmin` > `soporte` (detalle en `docs/specs/CU-01-autenticacion-usuarios.md`). El rol viaja en el JWT. El backend valida el rol en **cada** endpoint
(middleware + chequeo en service). El frontend solo oculta acciones: nunca es la única barrera.

## 7. Reglas anti-alucinación (cumplir siempre)

1. Antes de agregar una dependencia, leé `go.mod` / `package.json` / `pyproject.toml`. Si no está, **proponela en el plan**, no la instales sola.
2. Antes de crear o consumir un endpoint, verificá que exista en `api/openapi.yaml`. Si no existe, primero se modifica el contrato.
3. Antes de escribir SQL, leé las migraciones en `backend/db/migrations/`. No inventes columnas.
4. Para APIs de librerías, consultá documentación actualizada con el MCP `context7` en lugar de confiar en tu memoria.
5. Si un cambio toca más de 2 archivos, escribí primero un plan breve (archivos, dependencias, migraciones, riesgos).
6. Si no estás seguro de algo, preguntá. No inventes rutas, nombres de tablas ni variables de entorno.
7. No edites código generado a mano (`backend/internal/api/gen/`, `backend/internal/db/gen/`, `frontend/src/api/gen/`). Se regenera.

## 8. Flujo de trabajo (Spec-Driven Development)

Por cada caso de uso:

1. **Spec**: `docs/specs/CU-XX-*.md` completa y aprobada.
2. **Contrato**: endpoints y schemas en `api/openapi.yaml`.
3. **Plan** (modo Plan, solo lectura): archivos, migraciones, dependencias, tests.
4. **Build**: regenerar código (`make generate`), migración, query sqlc, service, handler, UI.
5. **Loop de autocorrección**: correr la verificación del servicio afectado; si falla, leer el error, corregir, repetir hasta verde.
6. **Cierre**: actualizar `docs/roadmap.md`, registrar el prompt clave en `docs/prompts/`, redactar descripción del PR desde `git diff`.

## 9. Comandos de verificación

```bash
# Raíz
make generate          # oapi-codegen + sqlc + orval
make verify            # corre los tres bloques de abajo
docker compose up -d --build

# Backend
cd backend && go vet ./... && golangci-lint run && go test ./...

# Frontend
cd frontend && npm run lint && npm run typecheck && npm run test -- --run

# AI service
cd ai-service && ruff check . && mypy . && pytest
```

Un cambio **no está terminado** hasta que la verificación del servicio afectado pasa.

## 10. Convenciones generales

- Idioma: código, identificadores y commits en **inglés**; documentación, UI y respuestas de OPLBot en **español**.
- Commits: Conventional Commits (`feat(backend): add server inventory endpoints`). Un caso de uso = una rama `feat/cu-XX-nombre`.
- Fechas: UTC en base y API (RFC 3339); se formatean a hora local solo en el frontend.
- IDs: UUID v4.
- Baja lógica (`deleted_at`) para servidores, servicios y documentos.
- Notificaciones solo internas (tabla `notifications`); no hay email ni Telegram por ahora.
- Todo debe poder correr en Debian sin dependencias de servicios cloud: nada de SaaS obligatorios.

## 11. Seguridad (no negociable)

- Nunca leas ni escribas `.env`, claves (`*.pem`, `id_rsa*`) ni dumps. Usá `.env.example` como referencia.
- Nunca escribas secretos en código, `opencode.json` ni docs. Todo secreto viene de variables de entorno.
- Solo datos ficticios: IPs de rangos de documentación (`192.0.2.0/24`, `198.51.100.0/24`, `203.0.113.0/24`), dominios `*.example.com` / `*.lab.local`.
- OPLBot y sus tools son de **solo lectura**. Ninguna tool del agente puede hacer POST/PUT/PATCH/DELETE.
- OPL no ejecuta comandos remotos ni guarda credenciales de servidores.

## 12. MCP disponibles

- `context7`: documentación actualizada de librerías. Usalo antes de escribir código con una API que no viste en el repo.
- `github`: issues (uno por CU), ramas, PRs y estado del CI.
- `filesystem`: limitado a este repositorio.
- `memory`: registrá decisiones de arquitectura importantes y consultalas al iniciar una sesión.
