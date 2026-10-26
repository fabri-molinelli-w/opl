# CU-08 · Notificaciones internas

| Campo | Valor |
|---|---|
| Actor principal | Sistema |
| Prioridad | Media |
| Depende de | CU-03, CU-10, CU-06 |

## Descripción
Notificaciones **dentro de la aplicación**: una campana en la barra superior con contador de no leídas
y un panel con el listado. No se usan canales externos.

## Eventos
| Evento | Disparador | Destinatarios |
|---|---|---|
| `service.down` | servicio pasa a DOWN | sysadmin, admin |
| `incident.opened` | incidente SEV1 o SEV2 abierto | todos |
| `incident.assigned` | me asignan un incidente | el asignado |
| `certificate.critical` | certificado pasa a critical | sysadmin, admin *(pendiente, CU-06)* |
| `document.updated` | se edita un documento que yo creé | el autor original |

## Reglas de negocio
- RN-1 Cada notificación pertenece a un usuario y tiene estado leída / no leída.
- RN-2 Anti-duplicados: misma `dedupe_key` (evento + entidad) no se repite al mismo usuario mientras haya una no leída.
- RN-3 *(Pendiente)* Cada usuario puede desactivar tipos de evento que no le interesan.
- RN-4 Las notificaciones se generan de forma asíncrona; un error no bloquea la operación que la disparó.
- RN-5 Retención: 30 días.

## Modelo de datos
- `notifications(id, user_id fk, event_type, title, body, link text, dedupe_key, read_at null, created_at)`
- *(Pendiente)* `notification_prefs(user_id fk, event_type, enabled bool, primary key(user_id, event_type))`
- Migración `00007`; RN-2 se garantiza con un índice único parcial `(user_id, dedupe_key) WHERE read_at IS NULL`.

## Endpoints
| Método | Ruta | Rol mínimo |
|---|---|---|
| GET | /api/me/notifications?unread=true | soporte |
| GET | /api/me/notifications/unread-count | soporte |
| POST | /api/me/notifications/{id}/read | soporte |
| POST | /api/me/notifications/read-all | soporte |
| GET / PUT | /api/me/notification-prefs | soporte *(pendiente)* |

## UI
- Campana con contador; se actualiza cada 30 s.
- Panel desplegable con las últimas 20; click en una notificación lleva al recurso (`link`) y la marca leída.

## Criterios de aceptación
- [ ] CA-1 Al simular una caída, los usuarios sysadmin ven la notificación sin recargar la página.
- [ ] CA-2 Durante la misma caída no se genera otra notificación igual.
- [ ] CA-3 "Marcar todas como leídas" deja el contador en 0.

## Fuera de alcance (por ahora)
Email por SMTP (posible a futuro con `net/smtp` de Go), Telegram, push del navegador.
