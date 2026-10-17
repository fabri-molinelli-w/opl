---
description: Implementar un caso de uso ya planificado. Uso /implementar-cu CU-02
agent: build
---
Implementá el caso de uso $ARGUMENTS siguiendo el plan aprobado en esta sesión.

Orden obligatorio:
1. Contrato: modificá api/openapi.yaml primero.
2. Migración goose + queries sqlc.
3. `make generate` (oapi-codegen, sqlc, orval). No edites archivos dentro de carpetas gen/.
4. Service (lógica de negocio) con sus tests unitarios.
5. Handler (solo traducción HTTP ↔ service).
6. UI en frontend/src/features/.
7. Loop de autocorrección: ejecutá la verificación de cada servicio que tocaste
   (ver AGENTS.md, sección 9). Si algo falla, leé el error, corregí y repetí hasta que todo pase.

No agregues dependencias que no estén en el plan aprobado. Si aparece algo no previsto, frená y preguntame.
Al terminar, mostrame la lista de archivos cambiados y el resultado final de la verificación.
