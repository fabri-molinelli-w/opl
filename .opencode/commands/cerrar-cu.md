---
description: Cerrar un caso de uso - roadmap, bitácora y descripción del PR. Uso /cerrar-cu CU-02
agent: build
---
Cambios realizados en esta rama:

!`git diff --stat main...HEAD`

!`git log --oneline main..HEAD`

Para el caso de uso $ARGUMENTS:

1. Verificá, uno por uno, los criterios de aceptación de su spec en docs/specs/ e indicá qué test cubre cada uno.
2. Marcá el CU como ✅ en docs/roadmap.md.
3. Creá docs/prompts/<número>-$ARGUMENTS.md usando la plantilla de docs/prompts/README.md (dejá en blanco lo que no sepas, yo lo completo).
4. Escribí PR_DESCRIPTION.md con: resumen, qué cambió y por qué (relación entre contrato, migraciones, services y UI), cómo probarlo, criterios de aceptación cubiertos.
5. Si se tomó alguna decisión de arquitectura nueva, agregala a docs/decisiones.md y guardala en el MCP memory.
