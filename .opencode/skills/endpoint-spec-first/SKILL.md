---
name: endpoint-spec-first
description: Pasos obligatorios para agregar o modificar un endpoint REST del backend Go de OPL, empezando por api/openapi.yaml y terminando con tests y cliente regenerado. Usar siempre que se cree o cambie un endpoint.
compatibility: opencode
metadata:
  opencode/slash: "true"
---

## Cuándo usarme
Siempre que haya que crear, modificar o eliminar un endpoint de la API de OPL.

## Pasos

1. **Contrato primero.** Editá `api/openapi.yaml`:
   - `operationId` en camelCase (`listServers`, `createIncident`).
   - Tag igual al dominio (`servers`, `incidents`, ...).
   - `security: [bearerAuth: []]` salvo login y webhook.
   - Request y response con schemas en `components/schemas` (nada inline salvo arrays triviales).
   - Declarar **todas** las respuestas de error posibles usando `$ref: '#/components/responses/...'` (BadRequest, Unauthorized, Forbidden, NotFound, Conflict).
   - Agregar la extensión `x-min-role: soporte|sysadmin|admin`.
2. **Regenerar:** `make generate`. Si falla, el YAML está mal: corregilo antes de seguir.
3. **Service:** método nuevo en `backend/internal/service/<dominio>.go` que recibe `context.Context` y tipos de dominio, y devuelve errores de `internal/domain`.
4. **Handler:** implementar el método de `StrictServerInterface` en `backend/internal/api/handlers/<dominio>.go`:
   - Chequear rol mínimo.
   - Convertir tipos generados ↔ tipos de dominio.
   - Delegar al service. Traducir errores con el helper central.
5. **Tests:** service con fakes + handler con `httptest` (caso feliz, 400, 403, 404/409 según aplique).
6. **Frontend:** usar el hook generado por Orval; no escribir fetch manual.
7. **Verificar:** `go vet ./... && golangci-lint run && go test ./...` y en frontend `npm run typecheck`.

## Errores comunes a evitar
- Escribir el handler antes que el contrato.
- SQL o lógica de negocio dentro del handler.
- Editar archivos dentro de `gen/`.
- Devolver `password_hash` o datos internos en un schema de respuesta.
