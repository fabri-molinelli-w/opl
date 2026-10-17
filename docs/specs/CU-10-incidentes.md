# CU-10 · Gestión de incidentes

| Campo | Valor |
|---|---|
| Actor principal | Soporte IT / SysAdmin |
| Prioridad | Alta |
| Depende de | CU-03 (incidentes automáticos), CU-07 (deploys relacionados, opcional) |

## Descripción
Registro y seguimiento de incidentes con severidad, responsable y línea de tiempo. Los incidentes se
pueden abrir a mano o automáticamente cuando un servicio falla N chequeos seguidos.

## Reglas de negocio
- RN-1 Severidad ∈ {SEV1, SEV2, SEV3, SEV4}.
- RN-2 Máquina de estados (cualquier otra transición → 409):
  - `open → in_progress`
  - `in_progress → resolved` (requiere `root_cause` no vacío)
  - `resolved → closed`
  - `resolved → in_progress` (reapertura)
- RN-3 Cada cambio de estado genera un `incident_event` de tipo `status_change` en la **misma transacción**.
- RN-4 Apertura automática: tras `HEALTHCHECK_FAILS_TO_INCIDENT` chequeos DOWN consecutivos, si el servicio no tiene un incidente no cerrado, se abre uno SEV2 con `auto_created = true` y título "Servicio <name> caído".
- RN-5 Si el servicio vuelve a UP, se agrega un evento `system` al incidente abierto ("El servicio volvió a responder"), pero **no** se resuelve solo.
- RN-6 Tiempo de resolución = `resolved_at - opened_at` (se muestra en minutos).
- RN-7 El detalle muestra los deploys del servicio en las 24 h previas a la apertura.

## Modelo de datos
- `incidents(id, title, description, service_id fk null, severity, status, assignee_id fk null, root_cause text null, auto_created bool, opened_at, resolved_at null, closed_at null, created_at, updated_at)`
- `incident_events(id, incident_id fk, author_id fk null, kind, body text, created_at)`

## Endpoints
| Método | Ruta | Rol mínimo | Notas |
|---|---|---|---|
| GET | /api/incidents | soporte | filtros `status`, `severity`, `service_id` |
| POST | /api/incidents | soporte | |
| GET | /api/incidents/{id} | soporte | incluye eventos y deploys relacionados |
| PATCH | /api/incidents/{id} | soporte | título, severidad, responsable |
| POST | /api/incidents/{id}/transition | soporte | `{to, root_cause?}` → 200 / 409 |
| POST | /api/incidents/{id}/events | soporte | nota en la línea de tiempo |

## UI
- Listado con filtros y badge de severidad.
- Detalle: línea de tiempo, botones de transición válidos según estado, deploys recientes, botón "Generar post-mortem" (usa OPLBot, CU-09).

## Criterios de aceptación
- [ ] CA-1 No se puede pasar a `resolved` sin causa raíz (409).
- [ ] CA-2 `open → closed` directo devuelve 409.
- [ ] CA-3 Apagar un servicio del lab abre un incidente automático tras N fallos.
- [ ] CA-4 Al resolver, se calcula y muestra el tiempo de resolución.
