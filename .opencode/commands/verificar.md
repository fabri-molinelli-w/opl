---
description: Correr el loop de autocorrección completo (lint + types + tests)
agent: build
---
Estado actual del repo:

!`git status --short`

Ejecutá la verificación de cada servicio que tenga cambios (backend, frontend, ai-service) usando los
comandos de la sección 9 del AGENTS.md raíz.

Por cada fallo:
1. Mostrá el error.
2. Explicá la causa en una línea.
3. Corregí y volvé a correr la verificación de ese servicio.

Repetí hasta que todo pase o hasta 5 iteraciones. Si a la quinta sigue fallando, frená y explicame qué
está bloqueando. No desactives reglas del linter ni borres tests para que pase.
