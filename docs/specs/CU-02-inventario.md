# CU-02 · Gestión del inventario de servidores y servicios

| Campo | Valor |
|---|---|
| Actor principal | SysAdmin |
| Prioridad | Alta |
| Depende de | CU-01 |

## Descripción
ABM y consulta de servidores y de los servicios que corren en ellos, con búsqueda y filtros combinados.

## Reglas de negocio
- RN-1 `hostname` único entre servidores no eliminados (409).
- RN-2 `ip` debe ser IPv4 o IPv6 válida (400).
- RN-3 `environment` ∈ {prod, staging, dev}; `location` ∈ {onprem, cloud}.
- RN-4 `port` entre 1 y 65535.
- RN-5 Servicio: `type` ∈ {web, db, queue, other}; `check_type` ∈ {http, tcp}. Si es `http`, `check_target` es una URL válida y `expected_status` (default 200).
- RN-6 Baja lógica (`deleted_at`). Al dar de baja un servidor se dan de baja sus servicios.
- RN-7 Al crear un servicio, `current_status` = `UNKNOWN`.
- RN-8 Listados paginados (`page`, `page_size` default 20, máx 100).

## Modelo de datos
- `servers(id, hostname, ip inet, environment, os, location, owner, tags text[], created_at, updated_at, deleted_at)`
- `services(id, server_id fk, name, type, port int, check_type, check_target, expected_status int default 200, interval_seconds int default 30, latency_threshold_ms int default 1000, current_status text default 'UNKNOWN', created_at, updated_at, deleted_at)`

## Endpoints
| Método | Ruta | Rol mínimo | Notas |
|---|---|---|---|
| GET | /api/servers | soporte | filtros `q`, `environment`, `os`, `tag`, `status` |
| POST | /api/servers | sysadmin | 201, 400, 409 |
| GET | /api/servers/{id} | soporte | incluye sus servicios |
| PATCH | /api/servers/{id} | sysadmin | |
| DELETE | /api/servers/{id} | sysadmin | 204, baja lógica |
| GET | /api/services | soporte | filtros `server_id`, `environment`, `status`, `type` |
| POST | /api/services | sysadmin | |
| GET | /api/services/{id} | soporte | |
| PATCH | /api/services/{id} | sysadmin | |
| DELETE | /api/services/{id} | sysadmin | 204 |

## UI
- "Inventario": tabla de servidores con buscador y filtros combinables; ficha del servidor con sus servicios.
- Formularios con validación inline y mensajes claros del backend.

## Criterios de aceptación
- [ ] CA-1 Los filtros se combinan (ej. `environment=prod&tag=web`).
- [ ] CA-2 IP inválida → 400 con `details[].field = "ip"`.
- [ ] CA-3 Hostname repetido → 409.
- [ ] CA-4 Un servidor dado de baja no aparece en listados.

## Datos semilla (ficticios)
5 servidores (`web-01.lab.local`, `db-01.lab.local`, ...) con IPs `192.0.2.x`, apuntando a los contenedores del lab.
