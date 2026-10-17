# CU-06 · Control de vencimiento de certificados SSL/TLS

| Campo | Valor |
|---|---|
| Actor principal | Sistema; SysAdmin da de alta dominios |
| Prioridad | Media |
| Depende de | CU-02 |

## Descripción
El sistema obtiene el certificado de cada dominio/endpoint HTTPS registrado y clasifica su vencimiento.

## Reglas de negocio
- RN-1 Estado: `ok` (> 30 días), `warning` (≤ 30), `critical` (≤ 7 o vencido).
- RN-2 Chequeo diario con robfig/cron (03:00 UTC) y a demanda.
- RN-3 Se usa `tls.Dial` con `InsecureSkipVerify: true` **solo para leer** el certificado (el lab usa autofirmados); se guarda emisor y `not_after`.
- RN-4 Se puede crear un certificado a partir de un servicio HTTPS existente.
- RN-5 Cambio a `critical` emite evento `certificate.critical` (CU-08).

## Modelo de datos
`certificates(id, domain, port int default 443, service_id fk null, issuer, not_after timestamptz, days_left int, status, last_checked_at, error text null, created_at, updated_at)`

## Endpoints
| Método | Ruta | Rol mínimo |
|---|---|---|
| GET | /api/certificates?status= | soporte |
| POST | /api/certificates | sysadmin |
| POST | /api/certificates/{id}/check | sysadmin |
| DELETE | /api/certificates/{id} | sysadmin |

## UI
Listado ordenado por días restantes con indicador de color.

## Criterios de aceptación
- [ ] CA-1 El certificado autofirmado del lab (vence en pocos días) aparece en `warning` o `critical` y se refleja en el dashboard.
