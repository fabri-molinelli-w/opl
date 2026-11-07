# Bitácora de prompts

Registrar acá los prompts que produjeron cambios relevantes. Un archivo por caso de uso o tarea:
`docs/prompts/NN-descripcion.md` (por ejemplo `01-esqueleto-backend.md`).

## Plantilla

```markdown
# NN · Título

- **Fecha:**
- **Caso de uso / tarea:**
- **Herramienta y modo:** OpenCode · Plan | Build
- **Modelo:** (nombre exacto del modelo gratuito de OpenCode Zen usado)
- **MCP usados:** context7 / github / filesystem / memory

## Prompt

(texto exacto)

## Resultado

- Qué generó / qué archivos tocó.
- Iteraciones del loop de autocorrección (cuántas, qué errores aparecieron).

## Correcciones humanas

- Qué tuve que corregir o rechazar del plan y por qué (matriz de auditoría: acoplamiento,
  inyección de dependencias, gobernanza de tipos, transacciones).

## Aprendizaje

- Qué cambiaría en AGENTS.md / spec / prompt para la próxima.
```

## Modelos usados por etapa

| Etapa | Modelo Plan | Modelo Build | Notas |
|---|---|---|---|
| Fase 0 | | | |
| CU-01 | | | |
| CU-02 | | | OpenCode Zen |
| CU-05, CU-09, CU-03, CU-10, CU-04, CU-08 | Claude Opus (`claude-opus-5-5`) | Claude Opus | Claude Code (ADR-021), bitácoras 09–13 |
| CU-06, CU-07, Docker, Fase 5 | Claude Opus (`claude-opus-5-5`) | Claude Opus | Claude Code, bitácora 14 |
