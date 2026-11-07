# 14 · Cierre del TP: CU-06, CU-07, Docker y producción

- **Fecha:** 2026-11-07
- **Caso de uso / tarea:** CU-06 certificados, CU-07 deploys, ajustes de CU-02, Docker Compose completo y Fase 5.
- **Herramienta y modo:** Claude Code (ADR-021) · build directo (alcance mínimo, día de entrega).
- **Modelo:** Claude Opus (`claude-opus-5-5`)
- **MCP usados:** ninguno en esta sesión.

## Prompt

> La entrega final es hoy. Terminá todo lo que falta del proyecto: casos de uso pendientes y ajustes.
> Lo mínimo e indispensable, sin lujos, que quede listo para entregar.

## Resultado

- **CU-02:** IP única entre servidores activos (409) y SO obligatorio (400), validados en el servicio.
- **CU-06:** migraciones `00008`/`00009`, lectura TLS de solo lectura, estados ok/warning/critical, chequeo diario
  desde el scheduler (ADR-026), evento `certificate.critical` en la campana, pantalla y tarjeta del dashboard,
  laboratorio `lab-tls` con certificado autofirmado de 20 días.
- **CU-07:** alta manual, webhook con `X-Deploy-Token` (tiempo constante, público en la política), historial con
  filtros, deploys de las 24 h previas en el detalle del incidente, `scripts/demo-deploy-webhook.sh` y workflow
  `deploy-notify.yml`.
- **OPSBot:** tools `certificados_por_vencer` y `deploys_recientes` (GET); el test de solo lectura las incluye.
- **Docker:** Dockerfiles de backend (distroless), frontend (nginx con proxy `/api`) y ai-service; perfil `app`
  en `docker-compose.yml`; `docker-compose.prod.yml`; job `docker` en el CI.
- **Fase 5:** `docs/instalacion-debian.md` (HTTPS con Caddy o nginx), `scripts/backup-postgres.sh` + restauración.
- Nombre unificado: **OPSBot** en toda la documentación.

## Loops de autocorrección

- oapi-codegen renombró los enums de `Health` (`gen.Ok` → `gen.HealthStatusOk`) al aparecer otro enum con `ok`
  (`CertificateStatus`); el build lo marcó y se ajustó el handler.
- misspell marcó "intencional" en un comentario; se reescribió.
- Vitest: el mock de `fetch` del detalle de incidente devolvía un objeto para `/api/deploys`; se protegió el
  listado con `Array.isArray` en lugar de asumir la forma.
- Prettier con la configuración por defecto cambiaba el estilo de archivos existentes (punto y coma); se
  revirtió y solo se formatearon los archivos nuevos con el estilo del repo.

## Verificación

- Backend: `go vet`, `golangci-lint` (0 issues), unitarios e integración contra PostgreSQL real (servidor TLS de
  prueba, webhook, filtros de deploys, IP duplicada).
- Frontend: lint, typecheck, 30 tests, build. ai-service: ruff, mypy --strict, 14 tests.
- Docker: el entorno del agente no tiene Docker; las imágenes y los compose se validan en el job `docker` del CI.
