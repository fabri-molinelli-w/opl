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
