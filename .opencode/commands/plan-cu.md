---
description: Planificar un caso de uso (solo lectura). Uso /plan-cu CU-02
agent: plan
---
Vamos a planificar el caso de uso $ARGUMENTS.

1. Leé la spec correspondiente en docs/specs/ (el archivo que empieza con $ARGUMENTS).
2. Leé api/openapi.yaml, las migraciones existentes en backend/db/migrations/ y docs/roadmap.md.
3. Revisá go.mod, frontend/package.json y ai-service/pyproject.toml antes de proponer dependencias.
4. Si necesitás documentación de una librería, usá context7.

Entregá un plan en Markdown con estas secciones, sin escribir código:
- **Cambios al contrato OpenAPI** (paths y schemas nuevos o modificados, en YAML).
- **Migraciones** (nombre del archivo y SQL Up/Down).
- **Queries sqlc** nuevas.
- **Archivos a crear / modificar** por servicio, con una línea de qué hace cada uno.
- **Dependencias nuevas** (si las hay) y por qué.
- **Tests** que cubren cada criterio de aceptación (CA-n → nombre del test).
- **Riesgos y dudas** que necesito resolver antes de pasar a Build.

Revisá el plan contra la matriz de auditoría: acoplamiento (handler sin lógica ni SQL), inyección de
dependencias por constructor, gobernanza de tipos (sin any, tipos generados del contrato) y transacciones.
