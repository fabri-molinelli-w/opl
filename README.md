# OPL — Operations Pilot Library

[![CI](https://github.com/fabri-molinelli-w/opl/actions/workflows/ci.yml/badge.svg)](https://github.com/fabri-molinelli-w/opl/actions/workflows/ci.yml)

Portal de operaciones de infraestructura con asistente IA (OPLBot).
Trabajo Práctico Integrador · Desarrollo de Software y Arquitectura con IA · Grupo 17

| Integrante | Mail | GitHub |
|---|---|---|
| Fabrizio Molinelli Wallaces | fabrimolinelli6@gmail.com | fabri-molinelli-w |

## Qué hace

Pensado para equipos de **Soporte IT y SysAdmin**. Centraliza la documentación del área (procedimientos,
runbooks, guías), el inventario de servidores y servicios, un dashboard, health checks, incidentes,
certificados SSL/TLS, deploys y notificaciones internas. OPLBot responde consultas operativas combinando los
documentos (RAG sobre ChromaDB) con el estado actual del sistema (tools de solo lectura).

- Casos de uso: [docs/specs](docs/specs)
- Arquitectura: [docs/arquitectura.md](docs/arquitectura.md)
- Estado: [docs/roadmap.md](docs/roadmap.md)
- Decisiones: [docs/decisiones.md](docs/decisiones.md)

## Levantar el proyecto

```bash
cp .env.example .env     # completar valores
docker compose up -d --build
# Frontend: http://localhost:5173  ·  API: http://localhost:8080/api/health
```

## Demo end-to-end

1. `docker compose stop lab-nginx` → el servicio pasa a DOWN.
2. Tras N fallos se abre un incidente automático y aparece la notificación en la campana.
3. En OPLBot: "¿Hubo deploys recientes en este servicio y cómo hago el rollback?".
4. `docker compose start lab-nginx` → vuelve a UP. Resolver incidente con causa raíz.
5. "Generar post-mortem" desde el detalle del incidente.

---

## Metodología de AI Engineering

### Herramienta
Todo el desarrollo se hace con **OpenCode** usando modelos gratuitos de **OpenCode Zen**.
El modelo usado en cada etapa se registra en [docs/prompts/README.md](docs/prompts/README.md).

### Ingeniería de contexto
- `AGENTS.md` raíz: stack, estructura, arquitectura en capas, mapeo de errores HTTP, reglas anti-alucinación, comandos de verificación.
- `AGENTS.md` anidados en `backend/`, `frontend/`, `ai-service/`, cargados vía `instructions` en `opencode.json`.
- Skills en `.opencode/skills/` (`endpoint-spec-first`, `migracion-sqlc`).
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
| | | |
