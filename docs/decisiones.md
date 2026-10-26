# Registro de decisiones (ADR livianos)

Formato: una entrada por decisión. No se borran; si cambian, se agrega una nueva que la reemplaza.

---

## ADR-001 · Monorepo con tres servicios
- **Fecha:** 2026-10-17
- **Decisión:** frontend, backend y ai-service en un único repositorio.
- **Motivo:** un solo AGENTS.md raíz como contexto común; el contrato OpenAPI es compartido por los tres.

## ADR-002 · Spec-first con OpenAPI
- **Fecha:** 2026-10-17
- **Decisión:** `api/openapi.yaml` se edita antes que el código. oapi-codegen (strict) genera el servidor Go y Orval el cliente TS.
- **Motivo:** si el código no respeta el contrato, no compila. Reduce alucinaciones del agente en nombres de campos y códigos HTTP.

## ADR-003 · sqlc en lugar de un ORM
- **Fecha:** 2026-10-17
- **Decisión:** SQL escrito a mano + código Go tipado generado por sqlc, sobre pgx v5.
- **Motivo:** el agente ve el SQL real y los tipos generados; no puede inventar columnas sin que falle la generación.

## ADR-004 · OPLBot como servicio Python separado
- **Fecha:** 2026-10-17
- **Decisión:** FastAPI + LangChain en contenedor propio, sin puerto público. Tools de solo lectura que llaman al backend con el JWT del usuario.
- **Motivo:** LangChain es Python (requisito del curso) y así OPLBot hereda los permisos del usuario sin duplicar lógica.

## ADR-005 · Embeddings locales
- **Fecha:** 2026-10-17
- **Decisión:** sentence-transformers con `multilingual-e5-small` dentro del contenedor.
- **Motivo:** costo cero, runbooks en español, sin enviar el contenido a terceros para vectorizar.

## ADR-006 · Numeración de casos de uso
- **Fecha:** 2026-10-17
- **Decisión:** se usa la numeración de la presentación del proyecto OPL (PDF del grupo 17): CU-01 Auth, CU-02 Inventario, CU-03 Monitoreo de servicios e incidentes, CU-04 Dashboard, CU-05 Runbooks, CU-06 Certificados, CU-07 Deploys, CU-08 Notificaciones, CU-09 OPLBot, CU-10 Gestión de incidentes.
- **Motivo:** ese documento es la definición propia del proyecto; specs, issues y ramas usan la misma numeración.

## ADR-007 · MCP del proyecto
- **Fecha:** 2026-10-17
- **Decisión:** cuatro servidores MCP: Filesystem (local), Memory (local), GitHub (externo) y Context7 (externo).
- **Motivo:** son los definidos en la presentación del proyecto; cubren el mínimo de dos MCP con al menos uno externo.

## ADR-008 · Roles orientados a Soporte IT y SysAdmin
- **Fecha:** 2026-10-17
- **Decisión:** roles `admin`, `sysadmin` y `soporte` (jerárquicos). `soporte` puede crear y editar documentación e incidentes; `sysadmin` además gestiona inventario y monitoreo.
- **Motivo:** la aplicación se usará primero en un área de Soporte IT / SysAdmin. Otros perfiles (DevOps, lectura) se evaluarán más adelante.

## ADR-009 · Notificaciones solo internas
- **Fecha:** 2026-10-17
- **Decisión:** las notificaciones se guardan en la base y se muestran en una campana dentro de la app. Sin Telegram ni email por ahora.
- **Motivo:** simplicidad y sin dependencias externas. SMTP queda como extensión futura.

## ADR-010 · Documentación como foco de la primera etapa
- **Fecha:** 2026-10-17
- **Decisión:** el módulo de runbooks se amplía a una base de documentación con categorías, etiquetas, versiones y vínculos al inventario. Junto con el dashboard es la prioridad de la Fase 1.
- **Motivo:** es el problema más concreto del área (procedimientos dispersos en archivos Word) y la base que después usa OPLBot.

## ADR-011 · Producción en Debian con Docker Compose
- **Fecha:** 2026-10-17
- **Decisión:** el destino de producción es un servidor Debian de la empresa con Docker Compose, accesible solo desde la red interna.
- **Motivo:** el mismo `docker-compose` del desarrollo se reutiliza; no hay dependencias de servicios cloud obligatorios.

## ADR-012 · Node 24 LTS para el frontend
- **Fecha:** 2026-10-18
- **Decisión:** Node 24 LTS vía nvm en desarrollo e imagen node:24 en Docker.
- **Motivo:** Node 20 (el de Debian 13) está fuera de soporte desde abril 2026 y las versiones actuales de Orval y Vitest ya no lo soportan. Node 24 tiene soporte hasta abril 2028.

## ADR-013 · Autorización centralizada por operationID con default-deny
- **Fecha:** 2026-10-18
- **Decisión:** un único middleware de autorización (`internal/auth`) recibe el `operationID` que oapi-codegen (strict) pasa a cada `StrictMiddlewareFunc` y aplica una `Policy` que mapea cada operación de `api/openapi.yaml` a su mínimo rol (`x-min-role`). La allowlist pública (`getHealth`, `login`) es explícita y toda operación sin entrada se deniega (default-deny). El middleware revalida además el usuario en cada request.
- **Motivo:** evita chequeos de rol dispersos por handler y que un endpoint nuevo quede abierto por olvido; el contrato es el único lugar donde se declara el rol mínimo. La revalidación permite rechazar a un usuario desactivado aunque su token siga vigente (RN-3).

## ADR-014 · Sesión del frontend con JWT en sessionStorage
- **Fecha:** 2026-10-18
- **Decisión:** el frontend guarda el JWT y el usuario en `sessionStorage`. El cliente HTTP (`src/api/http.ts`) adjunta el token y, ante un 401 con token presente, limpia la sesión y redirige a `/login`. Al recargar la app, `AuthContext` revalida la sesión con `GET /api/auth/me`.
- **Motivo:** la sesión vive por pestaña y no persiste más de lo necesario (a diferencia de `localStorage`); la revalidación contra el backend hace que un usuario desactivado pierda el acceso sin esperar a que expire el token.

## ADR-015 · Inventario con integridad estructural en PostgreSQL
- **Fecha:** 2026-10-25
- **Decisión:** `servers.ip` se guarda como `inet`. Las lecturas usan `host(ip)` y las escrituras `$n::inet`; el dominio Go usa `string` y valida con `netip.ParseAddr`. El `hostname` es único sin distinguir mayúsculas mediante el índice parcial `UNIQUE (lower(hostname)) WHERE deleted_at IS NULL`. Los `tags` son `text[]` con índice GIN y se filtran con el operador de contención (`@>`).
- **Motivo:** la base garantiza integridad y unicidad aunque falle una validación de aplicación, sin agregar tablas auxiliares. El índice GIN mantiene eficientes los filtros combinados por etiqueta.

## ADR-016 · Baja lógica en cascada dentro de una transacción
- **Fecha:** 2026-10-25
- **Decisión:** la baja de un servidor ejecuta `SoftDeleteServer` y `SoftDeleteServicesByServer` en una única transacción `pgx.Tx`; si el servidor no existe, se hace rollback y se devuelve 404. Todas las lecturas filtran `deleted_at IS NULL`, y los listados de servicios excluyen además los servicios cuyo servidor fue dado de baja.
- **Motivo:** cumple RN-6 de forma atómica y evita servicios huérfanos visibles o bajas parciales ante un error.

## ADR-017 · Tests de integración opt-in contra PostgreSQL real
- **Fecha:** 2026-10-25
- **Decisión:** los tests que dependen de PostgreSQL usan el build tag `integration`, leen `TEST_DATABASE_URL` y se ejecutan con `make test-integration`, que crea la base `opl_test` si no existe. Quedan fuera del CI y del `go test ./...` habitual.
- **Motivo:** validar índices, `inet`, arrays/GIN, unicidad case-insensitive y transacciones que los fakes no pueden cubrir, sin exigir una base de datos en cada corrida local ni en CI.

## ADR-018 · Errores de dominio con detalles por campo
- **Fecha:** 2026-10-25
- **Decisión:** `domain.Error` incluye `Details []FieldError` (`field`, `issue`). Los services construyen esos detalles y el mapeo centralizado de handlers los serializa en el schema `Error.details` del contrato.
- **Motivo:** el frontend muestra la validación junto al campo sin interpretar mensajes libres, y se mantiene un único lugar de traducción dominio→HTTP.

## ADR-019 · Datos semilla como comando idempotente separado
- **Fecha:** 2026-10-25
- **Decisión:** los datos de demo viven en `backend/cmd/seed` y se ejecutan manualmente con `make seed`. El comando hace UPSERT de servidores por hostname y crea cada servicio solo si no existe. No se ejecuta en migraciones ni al arrancar el backend.
- **Motivo:** migraciones deterministas y ausencia de datos ficticios en producción, con una semilla reproducible para la demo que puede correrse varias veces sin duplicados.

## ADR-020 · Búsqueda de documentos con ILIKE en el MVP
- **Fecha:** 2026-10-26
- **Decisión:** la búsqueda de CU-05 usa `ILIKE` sobre título, contenido y etiquetas, escapando `%`, `_` y `\`. El orden se elige con `CASE` sobre valores de una lista cerrada (nunca SQL armado con texto del usuario).
- **Motivo:** alcanza para cientos de documentos y no agrega extensiones. Si la búsqueda se vuelve lenta, se migra a `pg_trgm` o a full-text search (`tsvector` en español).

## ADR-021 · Cambio de agente para el cierre del TP
- **Fecha:** 2026-10-26
- **Decisión:** a partir de CU-05 el desarrollo se hace con Claude Code (Claude Opus) en lugar de OpenCode, porque se agotaron los créditos de OpenCode Zen a dos días de la entrega.
- **Motivo:** mantener el ritmo. La metodología no cambia: AGENTS.md, spec → contrato → `make generate` → service → handler → UI, loops con `make verify`, plan por CU en `docs/plans/` y revisión humana en el punto de control.

## ADR-022 · Embeddings con fastembed (ONNX) en lugar de sentence-transformers
- **Fecha:** 2026-10-26
- **Decisión:** los embeddings se calculan localmente con `fastembed` y el modelo multilingüe
  `paraphrase-multilingual-MiniLM-L12-v2`. Reemplaza a sentence-transformers + `multilingual-e5-small` (ADR-005).
- **Motivo:** mismo principio (local, multilingüe, sin enviar contenido a terceros), pero sin PyTorch: la
  instalación pasa de varios GB a unos cientos de MB, clave para el servidor Debian. El modelo es configurable
  con `EMBEDDINGS_MODEL`.

## ADR-023 · Sincronización de la documentación por versión (pull)
- **Fecha:** 2026-10-26
- **Decisión:** el ai-service sincroniza Chroma antes de responder comparando `id` + `version` de los documentos
  contra lo indexado, con el JWT del usuario. No hay endpoint `/ingest` ni llamadas del backend al ai-service al
  guardar un documento.
- **Motivo:** el ai-service queda de solo lectura, no hay estado `index_status` que mantener y un documento
  editado se refleja en la siguiente consulta. Con cientos de documentos el costo es una consulta paginada.

