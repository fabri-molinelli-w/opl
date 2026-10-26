# OPL — Operations Pilot Library

[![CI](https://github.com/fabri-molinelli-w/opl/actions/workflows/ci.yml/badge.svg)](https://github.com/fabri-molinelli-w/opl/actions/workflows/ci.yml)

Portal de operaciones de infraestructura con asistente IA (OPSBot).
Trabajo Práctico Integrador · Desarrollo de Software y Arquitectura con IA · Grupo 17

| Integrante | Mail | GitHub |
|---|---|---|
| Fabrizio Molinelli Wallaces | fabrimolinelli6@gmail.com | fabri-molinelli-w |

## Qué hace

Pensado para equipos de **Soporte IT y SysAdmin**. Centraliza la documentación del área (procedimientos,
runbooks, guías), el inventario de servidores y servicios, un dashboard, health checks, incidentes,
incidentes y notificaciones internas. OPSBot responde consultas operativas combinando los documentos (RAG sobre
ChromaDB) con el estado actual del sistema (tools de solo lectura).

- Casos de uso: [docs/specs](docs/specs)
- Arquitectura: [docs/arquitectura.md](docs/arquitectura.md)
- Estado: [docs/roadmap.md](docs/roadmap.md)
- Decisiones: [docs/decisiones.md](docs/decisiones.md)

## Estado de los casos de uso

| CU | Caso de uso | Estado |
|---|---|---|
| CU-01 | Autenticación y usuarios con roles (admin / sysadmin / soporte) | ✅ |
| CU-02 | Inventario de servidores y servicios | ✅ |
| CU-03 | Monitoreo automático (health checks HTTP/TCP) | ✅ |
| CU-04 | Dashboard (inventario, servicios, incidentes, disponibilidad 24 h, documentación) | ✅ |
| CU-05 | Base de documentación (Markdown, búsqueda, versiones) | ✅ |
| CU-08 | Notificaciones internas (campana) | ✅ (sin preferencias por usuario) |
| CU-09 | OPSBot: chatbot con RAG + tools de solo lectura | ✅ |
| CU-10 | Gestión de incidentes (automáticos y manuales) | ✅ (sin deploys ni post-mortem) |
| CU-06 | Certificados SSL/TLS | ⬜ trabajo futuro |
| CU-07 | Deploys y webhook | ⬜ trabajo futuro |

## Levantar el proyecto (desarrollo local)

Requisitos: Docker, Go (versión de `backend/go.mod`), Node 24 (`frontend/.nvmrc`), Python ≥ 3.12.

```bash
cp .env.example .env                    # completar valores (ver comentarios)
docker compose up -d postgres lab-nginx # base + servicio de laboratorio
cd frontend && npm ci && cd ..
make setup-ai                           # venv del ai-service (una sola vez)

make run-backend                        # terminal 1 · API en http://localhost:8080 (aplica migraciones)
make seed                               # enseguida, una vez: datos ficticios + laboratorio + documentos
make run-ai                             # terminal 2 · OPSBot (necesita LLM_* y AI_INTERNAL_TOKEN en .env)
make run-frontend                       # terminal 3 · http://localhost:5173
```

- `make seed` es idempotente y solo para desarrollo. Deja el monitoreo pausado en los servidores ficticios
  (no existen) y activo en `lab-nginx`.
- OPSBot usa cualquier LLM con API compatible con OpenAI y *tool calling* (Gemini, Groq, OpenCode Zen, Ollama):
  se configura con `LLM_BASE_URL`, `LLM_API_KEY` y `LLM_MODEL`. La primera consulta descarga el modelo de
  embeddings local (~220 MB).
- Verificación completa: `make verify` (backend, frontend y ai-service) · `make test-integration` (PostgreSQL real).

## Demo end-to-end

1. Dashboard: tarjetas de servicios, incidentes, disponibilidad y documentación (se refresca cada 30 s).
2. `docker compose stop lab-nginx` → en la ficha de `lab-01.lab.local` el servicio pasa a DOWN.
3. Tras 3 chequeos fallidos se abre el incidente automático "Servicio lab-nginx caído" (SEV2) y aparece en la campana.
4. OPSBot: "¿Hay incidentes abiertos?" y "¿Cómo reinicio Nginx?" → respuesta con fuentes (incidente y documento).
5. `docker compose start lab-nginx` → vuelve a UP y la línea de tiempo registra la recuperación.
6. Resolver el incidente: sin causa raíz devuelve 409; con causa raíz muestra el tiempo de resolución.

---

## Metodología de AI Engineering

### Herramienta
La Fase 0, CU-01 y CU-02 se desarrollaron con **OpenCode** y modelos gratuitos de **OpenCode Zen** (modos
Plan/Build, MCP, comandos y skills de `.opencode/`). A dos días de la entrega se agotaron los créditos y el resto
de los casos de uso se hizo con **Claude Code** (ADR-021), con la misma metodología: specs, contrato primero,
código generado, plan por CU en `docs/plans/`, loops de autocorrección y revisión humana antes de cada merge.
La bitácora de prompts está en [docs/prompts/](docs/prompts/).

### Loops de autocorrección (ejemplos reales)
- `golangci-lint` (misspell) marcó palabras en español como errores de inglés → se configuraron excepciones en lugar de cambiar textos de la UI (bitácora 09).
- Un test de Vitest fallaba porque `userEvent.setup()` reemplaza `navigator.clipboard` (bitácora 09).
- La revisión del flujo de OPSBot detectó que el `WriteTimeout` de 10 s cortaba respuestas del LLM (bitácora 11).
- La revisión humana del plan de CU-05 exigió que el orden del listado no se arme concatenando SQL (inyección).

### CI (GitHub Actions)
Jobs `backend` (vet, golangci-lint, tests), `frontend` (lint, typecheck, tests, build), `ai-service` (ruff,
mypy --strict, pytest) y `contract` (regenera el código desde `api/openapi.yaml` y falla si hay diferencias).

### Ingeniería de contexto
- `AGENTS.md` raíz: stack, estructura, arquitectura en capas, mapeo de errores HTTP, reglas anti-alucinación, comandos de verificación.
- `AGENTS.md` anidados en `backend/`, `frontend/`, `ai-service/`, cargados vía `instructions` en `opencode.json`.
- Skills en `.opencode/skills/` (`endpoint-spec-first`, `migracion-sqlc`).
- Plan aprobado por caso de uso en `docs/plans/` (sobrevive a la compactación de contexto del agente).
- Comandos en `.opencode/commands/`: `/plan-cu`, `/implementar-cu`, `/verificar`, `/cerrar-cu`, `/nuevo-endpoint`.

### Spec-Driven Development
Spec del CU (`docs/specs`) → contrato (`api/openapi.yaml`) → código generado (oapi-codegen, sqlc, Orval) → lógica.

### Flujo por caso de uso
1. `/plan-cu CU-XX` en modo **Plan** (solo lectura).
2. Revisión humana con la matriz de auditoría (acoplamiento, inyección de dependencias, tipos, transacciones).
3. `/implementar-cu CU-XX` en modo **Build**.
4. `/verificar`: loop de autocorrección con go vet, golangci-lint, go test · ESLint, tsc, Vitest · Ruff, mypy, pytest.
5. `/cerrar-cu CU-XX`: roadmap, bitácora de prompts y descripción del PR.

### Seguridad y exclusión de contexto
- `.gitignore` + `.aisignore`. Como OpenCode no lee `.aisignore` de forma nativa, las mismas reglas se aplican como `deny` en `permission.read` de `opencode.json`.
- Edición de archivos y comandos de terminal requieren aprobación; `git push`, `rm -rf` y borrado de volúmenes están denegados.
- Solo datos ficticios (IPs de documentación, dominios `.example.com` / `.lab.local`).
- Secretos solo en variables de entorno.

## Servidores MCP

Declarados en [`opencode.json`](opencode.json).

| MCP | Tipo | Uso | Configuración |
|---|---|---|---|
| context7 | Externo (remoto) | Documentación actualizada de librerías | sin credenciales |
| github | Externo (remoto) | Issues por CU, ramas, PRs, estado del CI | `GITHUB_PAT` en el entorno (token fine-grained solo para este repo) |
| filesystem | Local | Lectura/escritura acotada al repo | ruta absoluta del proyecto |
| memory | Local | Decisiones de arquitectura entre sesiones | `.opencode/memory.jsonl` |

Comprobar: `opencode mcp list`.

## Modelos utilizados

| Etapa | Modelo | Notas |
|---|---|---|
| Fase 0, CU-01, CU-02 | Modelos gratuitos de OpenCode Zen | Ver cada entrada de `docs/prompts/` |
| CU-05 en adelante | Claude Opus (`claude-opus-5-5`) vía Claude Code | Se agotaron los créditos de OpenCode (ADR-021) |
