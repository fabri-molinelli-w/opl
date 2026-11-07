# CU-07 · Registro de deploys y cambios

| Campo | Valor |
|---|---|
| Actor principal | SysAdmin; pipeline de CI/CD |
| Prioridad | Media |
| Depende de | CU-02 |

## Descripción
Historial de despliegues cargados a mano o recibidos desde GitHub Actions por webhook.

## Reglas de negocio
- RN-1 `result` ∈ {success, failed, rollback}; `environment` ∈ {prod, staging, dev}.
- RN-2 El webhook `POST /api/deploys/webhook` se autentica con header `X-Deploy-Token` (= `DEPLOY_WEBHOOK_TOKEN`), no con JWT.
- RN-3 El webhook identifica el servicio por `service_name` + `environment`; si no existe → 404.
- RN-4 Comparación del token en tiempo constante (`subtle.ConstantTimeCompare`).

## Modelo de datos
`deploys(id, service_id fk, version, environment, author, result, notes, source text check in ('manual','webhook'), deployed_at, created_at)`

## Endpoints
| Método | Ruta | Auth |
|---|---|---|
| GET | /api/deploys?service_id=&environment=&from=&to= | soporte |
| POST | /api/deploys | sysadmin |
| POST | /api/deploys/webhook | X-Deploy-Token |

## Entregable extra
`.github/workflows/deploy-notify.yml` (manual, `workflow_dispatch`) hace el `POST .../api/deploys/webhook`. Para la demo local: `make demo-deploy` (`scripts/demo-deploy-webhook.sh`).

## Criterios de aceptación
- [x] CA-1 Un workflow de ejemplo registra un deploy de punta a punta (`deploy-notify.yml` / `make demo-deploy`).
- [x] CA-2 Token inválido → 401 (también si `DEPLOY_WEBHOOK_TOKEN` está vacío: webhook deshabilitado).
- [x] CA-3 Los deploys cercanos a un incidente aparecen en su detalle (24 h previas a la apertura).

## Implementación
- Migración `00008`. En el alta manual el `environment` se toma del servidor del servicio.
- El webhook es público en la política de autorización (no usa JWT) y valida el token en el servicio.
- Pantalla **Deploys** (filtros por servicio y ambiente, alta manual para sysadmin) y tool de OPSBot `deploys_recientes`.
