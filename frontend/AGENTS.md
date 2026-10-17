# AGENTS.md — frontend (React + TypeScript)

Aplica a todo lo que esté dentro de `/frontend`. Complementa el `AGENTS.md` raíz.

## Estructura

```
frontend/src/
├── api/
│   ├── gen/            GENERADO por Orval desde ../api/openapi.yaml — no editar
│   └── http.ts         Instancia fetch/axios con baseURL y header Authorization
├── auth/               Contexto de sesión, guardas de ruta por rol
├── components/         Componentes reutilizables (Button, Table, StatusBadge, ...)
├── features/           Una carpeta por dominio: servers/, documents/, dashboard/, incidents/, chat/ ...
│   └── servers/
│       ├── ServersPage.tsx
│       ├── ServerForm.tsx
│       └── servers.test.tsx
├── layouts/            AppLayout con sidebar + panel lateral de OPLBot
├── routes.tsx
└── main.tsx
```

## Reglas

- **TypeScript strict**. Prohibido `any`; si algo es desconocido usar `unknown` y estrecharlo.
- Los tipos de datos de la API **siempre** vienen de `src/api/gen/`. No redefinir a mano interfaces que ya existen en el contrato.
- Fetching y cache con **TanStack Query** (hooks generados por Orval). Nada de `useEffect` + `fetch` manual para datos de la API.
- Componentes funcionales, nombres en `PascalCase`, hooks en `useCamelCase`, archivos de componente `.tsx`.
- Estilos solo con Tailwind. Sin CSS inline ni archivos `.css` por componente (salvo `index.css`).
- Gráficos con Recharts (uptime %, latencia 24 h).
- Estados de servicio con colores consistentes: UP verde, DEGRADED ámbar, DOWN rojo, UNKNOWN gris.
- Mostrar/ocultar acciones según el rol del usuario, pero asumir siempre que el backend puede responder 403 y manejarlo.
- Manejo de errores: mostrar `message` del schema `Error`; en 401 redirigir a login.
- Textos de la UI en español. Fechas formateadas a hora local con `Intl.DateTimeFormat('es-AR')`.
- Refresco automático del dashboard con `refetchInterval` (10 s) de TanStack Query.

## Tests

- Vitest + Testing Library. Al menos un test por pantalla principal (renderiza, muestra datos mockeados, respeta rol).

## Verificación (loop de autocorrección)

```bash
npm run generate     # orval
npm run lint         # eslint
npm run typecheck    # tsc --noEmit
npm run test -- --run
```
