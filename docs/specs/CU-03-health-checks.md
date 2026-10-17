# CU-03 · Monitoreo automático de servicios y de incidentes

| Campo | Valor |
|---|---|
| Actor principal | Sistema (scheduler); SysAdmin configura |
| Prioridad | Alta |
| Depende de | CU-02 |

## Descripción
El backend ejecuta periódicamente chequeos HTTP/HTTPS o TCP sobre cada servicio y guarda el resultado.

## Reglas de negocio
- RN-1 Chequeo HTTP: GET a `check_target`; UP si el código == `expected_status` y la latencia ≤ `latency_threshold_ms`; DEGRADED si el código es correcto pero la latencia supera el umbral; DOWN si hay error, timeout o código distinto.
- RN-2 Chequeo TCP: conexión a `host:port`; UP si conecta, DOWN si no. DEGRADED por latencia igual que HTTP.
- RN-3 Timeout de cada chequeo: 5 s.
- RN-4 Cada servicio se chequea cada `interval_seconds` (mínimo 10).
- RN-5 Pool de workers concurrente (`HEALTHCHECK_WORKERS`); un chequeo lento no bloquea a los demás.
- RN-6 Cada resultado se guarda en `health_checks` y actualiza `services.current_status`.
- RN-7 Si cambia el estado, se emite un evento interno `service.status_changed` (lo consumen CU-10 y CU-08).
- RN-7b Tras `HEALTHCHECK_FAILS_TO_INCIDENT` chequeos DOWN consecutivos se abre un incidente automático (la regla completa está en CU-10, RN-4).
- RN-8 Chequeo manual a demanda desde la UI (sysadmin).
- RN-9 Retención del historial: 7 días (limpieza diaria).

## Modelo de datos
`health_checks(id, service_id fk, status, response_ms int null, http_status int null, error text null, checked_at timestamptz)` + índice `(service_id, checked_at desc)`.

## Endpoints
| Método | Ruta | Rol mínimo | Notas |
|---|---|---|---|
| GET | /api/services/{id}/checks | soporte | `from`, `to`, `limit` |
| POST | /api/services/{id}/checks/run | sysadmin | ejecuta ya y devuelve el resultado |

## UI
- En la ficha del servicio: estado actual, historial (tabla) y gráfico de latencia.
- Botón "Chequear ahora".

## Criterios de aceptación
- [ ] CA-1 Al detener un contenedor del lab (`docker compose stop lab-nginx`) el servicio pasa a DOWN en el siguiente ciclo.
- [ ] CA-2 Al volver a levantarlo regresa a UP.
- [ ] CA-3 El historial se puede consultar desde la ficha del servicio.
- [ ] CA-4 Test unitario del clasificador UP/DEGRADED/DOWN con un `httptest.Server`.
