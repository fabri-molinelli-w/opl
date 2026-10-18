# 06 · CU-01 Autenticación y gestión de usuarios (backend + frontend)

- **Fecha:** 2026-10-18
- **Caso de uso / tarea:** CU-01 — Autenticación y gestión de usuarios con roles. Backend (Go) ya terminado y verificado; cierre del frontend (React + TypeScript) según `docs/plans/CU-01.md`.
- **Herramienta y modo:** OpenCode · Build (con reconocimiento de estado previo)
- **Modelo:** opencode/deepseek-v4.1-flash-free
- **MCP usados:** filesystem

## Prompt

> Leé docs/plans/CU-01.md, AGENTS.md y frontend/AGENTS.md. El backend de CU-01 está terminado y verificado.
> Revisá qué partes del frontend del plan ya existen (git status y los archivos en frontend/src)
> y continuá solo con lo que falta.

## Resultado

- Alcance real: el backend de CU-01 (`internal/auth`, `internal/service/{auth,user}.go`,
  `internal/repository/user.go`, handlers + `db/migrations/00002_create_users.sql`,
  `db/queries/users.sql`) ya estaba escrito y verificado en la rama `feat/cu-01-auth`; el contrato
  (`api/openapi.yaml`) y el cliente generado (`frontend/src/api/gen/{auth,users,model}`) también.
  El trabajo de esta sesión fue **completar el frontend** y no tocar contrato, backend ni dependencias.
- Ya existían del plan: `api/gen/*`, `auth/{session,roles,redirect}.ts` y stubs de páginas.
- Creados:
  - `src/auth/context.ts`, `src/auth/AuthContext.tsx` (bootstrap con `getCurrentUser`), `src/auth/useAuth.ts`.
  - `src/auth/RequireAuth.tsx` y `src/auth/RequireRole.tsx` (guardas por sesión y por rol, con default-deny).
  - `src/features/auth/LoginPage.tsx`, `src/features/account/AccountPage.tsx` (cambio de contraseña),
    `src/features/users/UserForm.tsx` (alta/edición) y `src/features/users/UsersPage.tsx` real
    (tabla, alta, edición, cambio de rol y activar/desactivar).
- Editados:
  - `src/api/http.ts`: token desde `sessionStorage`; 401 con token → limpia sesión + redirige a `/login`;
    403 → mensaje `No tenés permisos`; se mantiene la excepción de 503 para el health.
  - `src/layouts/AppLayout.tsx`: menú filtrado por rol + usuario y logout en el header.
  - `src/routes.tsx`: `/login` público, `RequireAuth` envolviendo el layout y `RequireRole admin` en `/usuarios`.
  - `src/main.tsx`: `AuthProvider` dentro de `QueryClientProvider`.
  - `src/test/setup.ts`: `cleanup()` en `afterEach` (faltaba y acumulaba DOM entre tests).
- Tests nuevos: `src/api/http.test.ts` (401 limpia+redirige, 401 sin token no redirige, 403 → permisos),
  `src/features/auth/LoginPage.test.tsx` (login ok guarda token y navega; login ko muestra error y no guarda token)
  y `src/layouts/AppLayout.test.tsx` (menú por rol y datos del usuario).
- Iteraciones del loop de autocorrección: 2.
  1. `tsc --noEmit` falló en `UsersPage.tsx`: `data` es la unión `User[] | Error`; se resolvió estrechando
     por `response.status === 200` antes de usar `response.data`.
  2. `vitest` encontró múltiples coincidencias porque no había `cleanup()` entre tests; se agregó en el setup.
- Verificación final: `npm run lint`, `npm run typecheck` y `npm run test -- --run` en verde
  (4 archivos de test, 11 tests). Sin cambios en el contrato ni en el backend.

## Correcciones humanas

- (pendiente)

## Aprendizaje

- Conviene revisar `git status` y el árbol de `src` antes de escribir: parte del plan (helpers de sesión,
  roles y cliente generado) ya existía y sólo faltaba cablearla.
- El cliente Orval v8 devuelve uniones discriminadas por `status`; `data.data` queda como unión hasta
  estrechar por `status === 200`. Hay que estrechar en el componente, no castear.
- Testing Library no limpia el DOM automáticamente sin `globals: true` o `cleanup()` explícito: dos renders
  en el mismo archivo rompen las queries por duplicado. Mejor dejarlo en `src/test/setup.ts`.
- En `http.ts`, distinguir "401 con token" (sesión vencida → redirigir) de "401 sin token" (login fallido)
  evita que un login incorrecto dispare la redirección.
