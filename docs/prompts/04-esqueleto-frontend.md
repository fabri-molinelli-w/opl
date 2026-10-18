# 04 · Esqueleto frontend

- **Fecha:** 2026-10-18
- **Caso de uso / tarea:** 0.6 — Esqueleto frontend: Vite, Tailwind, Router, TanStack Query, Orval, layout con menú lateral
- **Herramienta y modo:** OpenCode · Plan + Build
- **Modelo:** opencode/deepseek-v4.1-flash-free
- **MCP usados:** context7, filesystem

## Prompt

> Leé AGENTS.md, frontend/AGENTS.md, api/openapi.yaml, Makefile y docs/roadmap.md.

> Tarea: paso 0.6 del roadmap. Esqueleto del frontend en /frontend.

> Alcance:
> 1. Proyecto React + TypeScript (strict) con Vite. Versiones actuales y estables, fijadas en package.json.
> 2. Tailwind CSS v4 con el plugin oficial de Vite (sin tailwind.config.js si v4 no lo necesita).
> 3. React Router y TanStack Query.
> 4. Orval: orval.config.ts que lea ../api/openapi.yaml y genere en src/api/gen/ hooks de TanStack Query
>    usando fetch. Un mutator en src/api/http.ts que centralice la base URL y deje preparado el header
>    Authorization (por ahora sin token).
> 5. Proxy de Vite: /api -> http://localhost:8080.
> 6. Layout con menú lateral: Dashboard, Documentación, Inventario, Usuarios. Solo Dashboard tiene contenido.
> 7. Página Dashboard: indicador "Estado del sistema" usando el hook generado de GET /api/health,
>    refresco cada 30 s, verde (ok/up), rojo (degraded/down o 503) o gris (sin conexión).
> 8. ESLint + Vitest + Testing Library, con un test del indicador.
> 9. Scripts: dev, build, generate, lint, typecheck, test.
> 10. Makefile raíz: agregar Orval a generate, lint+typecheck+test del frontend a verify, y target run-frontend.
>
> Sin librerías de componentes ni axios. El código generado por Orval se commitea y no se edita.
> Consultá con context7 la configuración actual de Vite, Tailwind v4, Orval y TanStack Query.

> Restricciones adicionales del usuario:
> - Usá solo los nombres que Orval genere realmente: corré orval primero y leé src/api/gen antes de escribir componentes.
> - Usá el tipo Health generado; nada de unknown ni de adivinar campos.
> - Prohibido "as any", también en tests.
> - El mutator no debe lanzar error con 503: el indicador distingue 503 (rojo) de error de red (gris).
> - En el Makefile solo AGREGÁ cosas: no modifiques los comandos existentes de backend ni agregues ai-service.
> - Mostrame la salida real de make generate y make verify.

## Resultado

- Node 24 LTS (`.nvmrc`, `engines.node >= 24`). TypeScript fijado en 5.9.3 por el peer de
  typescript-eslint (`<6.1.0`); TS 7 rompería el lint tipado.
- Orval `tags-split`: genera `src/api/gen/system/system.ts` (hook `useGetHealth`), `src/api/gen/model/*`
  (tipo `Health`, `HealthStatus`, `HealthDatabase`) e `index.ts` con barrel.
- El cliente fetch de Orval v8 devuelve un envelope `{ data, status, headers }`; el mutator
  `src/api/http.ts` se ajustó a esa firma y **no lanza en 503** (devuelve el envelope con status 503).
- Frontend verificado: `npm run lint`, `npm run typecheck`, `npm run test -- --run` (3 tests) y
  `npm run build` en verde.
- Makefile: 1 línea agregada a `generate`, 3 a `verify`, target `run-frontend` nuevo. Los comandos
  de backend quedaron intactos.

## Correcciones humanas

- (pendiente)

## Aprendizaje

- Orval v8 con `httpClient: 'fetch'` + mutator propio espera que el mutator devuelva el envelope
  `{ data, status, headers }`, no solo el body. Conviene leer el generado antes de escribir el mutator.
- En `tags-split` el archivo por tag se llama como el tag (`system/system.ts`), no como el `target`.
