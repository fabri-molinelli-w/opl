# CU-01 · Autenticación y gestión de usuarios con roles

| Campo | Valor |
|---|---|
| Actor principal | Administrador (login: todos) |
| Prioridad | Alta |
| Depende de | Fase 0 |

## Descripción
Los usuarios inician sesión con email y contraseña y reciben un JWT con su rol. El administrador
crea, edita, desactiva usuarios y cambia su rol.

## Roles
La aplicación está orientada a los equipos de **Soporte IT** y **SysAdmin**.

| Rol | Quién es | Puede |
|---|---|---|
| `admin` | Responsable de la plataforma | Todo lo de `sysadmin` + gestión de usuarios y configuración |
| `sysadmin` | Administración de servidores | Todo lo de `soporte` + ABM de inventario, health checks, certificados y deploys |
| `soporte` | Soporte IT / mesa de ayuda | Ver todo el sistema; crear y editar **documentación**; crear y gestionar **incidentes**; usar OPSBot |

Jerarquía: `admin` > `sysadmin` > `soporte`. Un rol superior puede todo lo que puede uno inferior.

## Reglas de negocio
- RN-1 Contraseñas con bcrypt (cost 12). Mínimo 8 caracteres.
- RN-2 JWT HS256 con claims `sub` (user id), `email`, `role`, `exp` (TTL `JWT_TTL`).
- RN-3 Un usuario desactivado no puede iniciar sesión (401) y sus tokens vigentes se rechazan.
- RN-4 Al arrancar, si no existe ningún admin, se crea uno con `ADMIN_EMAIL` / `ADMIN_PASSWORD`.
- RN-5 Un admin no puede desactivarse ni quitarse el rol de admin a sí mismo (409).
- RN-6 Email único (409 si se repite).
- RN-7 Cada usuario puede cambiar su propia contraseña (requiere la actual).

## Modelo de datos
`users(id uuid, email text unique, name text, password_hash text, role text check in ('admin','sysadmin','soporte'), active bool default true, created_at, updated_at)`

## Endpoints
| Método | Ruta | Rol mínimo | Respuestas |
|---|---|---|---|
| POST | /api/auth/login | público | 200 `{token, user}`, 400, 401 |
| GET | /api/auth/me | soporte | 200 `User`, 401 |
| POST | /api/auth/me/password | soporte | 204, 400, 401 |
| GET | /api/users | admin | 200 `User[]`, 401, 403 |
| POST | /api/users | admin | 201 `User`, 400, 403, 409 |
| PATCH | /api/users/{id} | admin | 200 `User`, 400, 403, 404, 409 |

`User` nunca incluye `password_hash`.

## UI
- Pantalla de login.
- El menú lateral muestra solo las secciones habilitadas para el rol.
- Pantalla "Usuarios" (solo admin): tabla, alta, cambio de rol, activar/desactivar.
- "Mi cuenta": cambio de contraseña.

## Criterios de aceptación
- [x] CA-1 Login correcto devuelve token; incorrecto devuelve 401 sin indicar si falló el email o la contraseña.
- [ ] CA-2 Un usuario `soporte` que intenta crear un servidor recibe 403 y no ve el botón.
      *(diferido a CU-02: aún no existe el endpoint de servidores. El control de rol y el ocultamiento del botón
      se cubren de forma genérica en `middleware_test.go::TestProtectedEndpoints_AuthMatrix` y `AppLayout.test.tsx`.)*
- [ ] CA-3 Un usuario `soporte` puede crear y editar documentación.
      *(diferido a CU-05: aún no existe el módulo de documentación.)*
- [x] CA-4 La contraseña en base está hasheada con bcrypt.
- [x] CA-5 Un usuario desactivado no puede loguearse.

## Fuera de alcance (por ahora)
Login con Active Directory / LDAP, recuperación de contraseña por email, SSO.
