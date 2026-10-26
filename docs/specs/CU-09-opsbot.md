# CU-09 · Asistente inteligente OPLBot

| Campo | Valor |
|---|---|
| Actor principal | Todos los roles |
| Prioridad | Alta |
| Depende de | CU-05 (RAG) y endpoints GET de CU-02, 03, 06, 07, 10 |

## Descripción
Chat en panel lateral. Agente LangChain que combina RAG sobre runbooks con tools de consulta del estado
del sistema, y redacta borradores de post-mortem.

## Flujo
1. Frontend → `POST /api/chat` (backend) con `{sessionId, message}`.
2. Backend valida JWT y reenvía a ai-service `POST /chat` con `X-Internal-Token`, `Authorization: Bearer <JWT del usuario>` y `X-User-Id`.
3. El agente decide: buscar runbooks, llamar tools, o ambas.
4. Responde `{answer, sources[{type: document|server|service, id, title}]}`. Las fuentes las registran las tools (no las inventa el LLM).
5. Si el ai-service no está configurado o falla → 503 `unavailable` con mensaje claro.

## Tools (solo GET, con el JWT del usuario)
| Tool | Endpoint backend | Estado |
|---|---|---|
| buscar_documentacion | (Chroma local, sincronizado desde GET /api/documents) | ✅ |
| estado_servicios | GET /api/services?status=&environment= | ✅ |
| buscar_servidores | GET /api/servers?q=&environment= | ✅ |
| incidentes_abiertos / detalle_incidente | GET /api/incidents | cuando exista CU-10 |
| certificados_por_vencer | GET /api/certificates | cuando exista CU-06 |
| deploys_recientes | GET /api/deploys | cuando exista CU-07 |

## Sincronización de la documentación (ADR-023)
Antes de cada respuesta, el ai-service compara `id` + `version` de `GET /api/documents` (sin contenido) con lo
indexado en Chroma y descarga solo los documentos nuevos o modificados; borra los que ya no existen. No hay
endpoint `/ingest` ni escrituras desde el backend: el ai-service sigue siendo de solo lectura.

## Reglas
- RN-1 Temperatura 0–0.2. Modelo, base URL y key por env vars.
- RN-2 Toda respuesta cita fuentes. Sin contexto suficiente → lo dice, no inventa.
- RN-3 Memoria por usuario + `sessionId` (últimos 10 turnos, en memoria del proceso). Una sesión nunca se comparte entre usuarios.
- RN-4 *(cuando exista CU-10)* `POST /api/incidents/{id}/postmortem-draft` → backend pide a ai-service un borrador en Markdown (resumen, línea de tiempo, causa raíz, acciones) a partir de los eventos. No se guarda automáticamente.
- RN-5 Ninguna tool modifica datos (test automatizado que lo verifica).

## Endpoints (backend)
| Método | Ruta | Rol mínimo |
|---|---|---|
| POST | /api/chat | soporte |
| POST | /api/incidents/{id}/postmortem-draft | soporte (pendiente, CU-10) |

## Preguntas de demo
- "¿Qué servicios están caídos en producción?"
- "¿Hubo deploys recientes en api-demo y cómo hago el rollback?"
- "¿Cómo restauro el backup de la base de staging?"

## Criterios de aceptación
- [ ] CA-1 Las respuestas indican su fuente.
- [ ] CA-2 Pregunta sin información en runbooks ni en el sistema → responde que no tiene información.
- [ ] CA-3 Un usuario `soporte` no obtiene datos que su rol no puede ver (ej. usuarios).
- [ ] CA-4 Ninguna tool hace POST/PUT/PATCH/DELETE.
