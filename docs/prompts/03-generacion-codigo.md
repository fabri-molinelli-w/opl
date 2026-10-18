# 01 · CÓdigo Backend

- **Fecha:** 2026-10-18
- **Caso de uso / tarea:** 0.5 - Generación: oapi-codegen + sqlc + Makefile (`make generate`, `make verify`)
- **Herramienta y modo:** OpenCode · Plan + Build
- **Modelo:** big-pickle (opencode/big-pickle)
- **MCP usados:** solo utilizo filesystem

## Prompt

> Leé AGENTS.md, backend/AGENTS.md, api/openapi.yaml, docs/roadmap.md y el código actual de backend/.

> Tarea: paso 0.5 del roadmap. Incorporar generación de código al backend.

> Alcance:
> 1. oapi-codegen v2: archivo backend/oapi-codegen.yaml que genere en backend/internal/api/gen/
   los modelos, la interfaz StrictServerInterface y el servidor para chi, a partir de ../api/openapi.yaml.
   Reemplazá el handler de /api/health escrito a mano por una implementación de StrictServerInterface
   en backend/internal/api/handlers/. El comportamiento debe ser el mismo.
> 2. sqlc: backend/sqlc.yaml con engine postgresql, driver pgx/v5, schema = db/migrations (formato goose),
   queries = db/queries, salida en backend/internal/db/gen/. Si sqlc exige al menos una query,
   agregá una mínima y explicá por qué.
> 3. Makefile en la raíz con los targets:
   - generate: corre oapi-codegen y sqlc (dejá un comentario indicando que Orval se agrega en el paso 0.6)
   - verify: go vet, golangci-lint run y go test dentro de backend/
   - run-backend: corre el backend apuntando a la base local
> 4. backend/.golangci.yml con una configuración mínima y razonable para golangci-lint v2.

> El código generado se commitea al repo y nunca se edita a mano.
> Consultá con context7 la configuración actual de oapi-codegen v2 (strict server + chi), sqlc y golangci-lint v2.

> No escribas código. Entregá: árbol de archivos nuevos/modificados, contenido propuesto de cada
> archivo de configuración, cambios en main.go y comandos de verificación.

> El test de health no debe depender de una base real ni saltearse. Reescribilo con un fake
> que implemente la interfaz que usa el handler (Ping que devuelve nil o error) y cubrí
> los dos casos: 200 ok/up y 503 degraded/down. Corré make verify hasta que pase. 

## Resultado

Cambios principales:
api/openapi.yaml                                    -   503 + components.responses.ServiceUnavailable en GET /api/health
backend/oapi-codegen.yaml                           -   chi-server + strict-server + models, skip-prune: true
backend/sqlc.yaml                                   -   postgresql, pgx/v5, schema goose → internal/db/gen/
backend/db/queries/system.sql                       -   query mínima Ping (sqlc falla sin queries)
backend/internal/api/gen/server.gen.go              -   generado, commiteado
backend/internal/db/gen/*.go                        -   generado, commiteado
backend/internal/api/handlers/health.go             -   StrictServerInterface: 200 ok/up o 503 degraded/down
backend/cmd/opl/main.go                             -   wiring api.HandlerFromMux(NewStrictHandler(...)) + patrón run() error (corrige hallazgos de errcheck/gocritic)
backend/.golangci.yml                               -   v2: standard + gocritic/misspell, exclusiones en gen/Makefile	generate, verify, run-backend

## Correcciones humanas

- Corrección en el Makefile: no estaba corriendo el run-backend porque no estaba modificado, aún seguía tomando credenciales a mano y esto suponia exponer credenciales en un archivo que pega en el repositorio.
La corrección ahora implica buscar dichas credenciales de postgres al archivo correspondiente.

## Aprendizaje

- Las limitaciones de Go respecto a las rutas que manipula
