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
- RN-2 Chequeo diario y a demanda. *Implementado:* el scheduler del backend (ADR-024) revisa cada hora y vuelve a chequear los certificados con más de 24 h desde el último chequeo (ADR-026; no se agregó robfig/cron).
- RN-3 Se usa `tls.Dial` con `InsecureSkipVerify: true` **solo para leer** el certificado (el lab usa autofirmados); se guarda emisor y `not_after`.
- RN-4 Se puede crear un certificado a partir de un servicio HTTPS existente. *Pendiente (extra):* hoy se carga dominio + puerto; `service_id` no se implementó.
- RN-5 Cambio a `critical` emite evento `certificate.critical` (CU-08).

## Modelo de datos
`certificates(id, domain, port int default 443, issuer, not_after timestamptz, days_left int, status, last_checked_at, error text null, created_at, updated_at)`

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
- [x] CA-1 El certificado autofirmado del lab (vence en pocos días) aparece en `warning` o `critical` y se refleja en el dashboard.

## Implementación
- Migraciones `00008` (tabla, único `lower(domain)+port`) y `00009` (evento `certificate.critical`).
- `service/release.go`: lectura TLS de solo lectura con timeout de 10 s; si falla queda `unknown` con el motivo.
- Laboratorio `lab-tls` en docker-compose: certificado autofirmado de `LAB_CERT_DAYS` días (20 → `warning`).
- Pantalla **Certificados**, tarjeta "Certificados por vencer" en el dashboard y tool de OPSBot `certificados_por_vencer`.
- Tests: regla de estados, alta con chequeo inmediato, dominio inalcanzable, evento crítico (unitarios) y
  servidor TLS real + base real (integración).
