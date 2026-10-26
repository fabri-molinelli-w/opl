# AGENTS.md — ai-service (Python, OPSBot)

Aplica a todo lo que esté dentro de `/ai-service`. Complementa el `AGENTS.md` raíz.

## Qué hace

Servicio interno (no expuesto a internet) que implementa OPSBot:
- `POST /chat`: recibe la consulta del usuario + su JWT y `X-User-Id` (reenviados por el backend) y devuelve
  respuesta con fuentes. Exige el header `X-Internal-Token` (= `AI_INTERNAL_TOKEN`).
- `GET /health`.
No hay `/ingest`: antes de responder se sincroniza Chroma por versión de documento (ADR-023).

## Estructura

```
ai-service/
├── app/
│   ├── main.py            FastAPI app, routers, dependencia de auth interna
│   ├── config.py          pydantic-settings: todas las env vars tipadas
│   ├── schemas.py         Modelos Pydantic de request/response
│   ├── rag/
│   │   ├── splitter.py    Chunking por encabezados Markdown + solapamiento
│   │   ├── embeddings.py  fastembed/ONNX multilingüe, local (ADR-022); HashEmbedder solo para tests
│   │   ├── store.py       Acceso a ChromaDB (colección "documents")
│   │   └── sync.py        Sincronización por versión contra GET /api/documents
│   ├── agent/
│   │   ├── prompt.py      System prompt de OPSBot
│   │   ├── tools.py       Tools de solo lectura (llaman GET al backend con el JWT del usuario)
│   │   ├── memory.py      Historial por session_id
│   │   └── agent.py       Armado del agente LangChain
│   └── backend_client.py  Cliente httpx hacia el backend Go
├── tests/
└── pyproject.toml
```

## Reglas de Python

- Python 3.12, **mypy --strict**, type hints en todas las funciones. Prohibido `Any` salvo en bordes con librerías sin tipos (comentarlo).
- Formato y lint con **Ruff**.
- Configuración solo vía `pydantic-settings`; nunca `os.environ[...]` disperso.
- I/O asíncrono (`async def`, `httpx.AsyncClient`).
- Antes de usar una API de LangChain, consultá **context7**: la API cambia seguido entre versiones.

## Reglas de OPSBot

- LLM: `ChatOpenAI(base_url=LLM_BASE_URL, api_key=LLM_API_KEY, model=LLM_MODEL, temperature=0.1)`. Nada hardcodeado.
- Tools (todas **GET**, todas reciben el JWT del usuario): `buscar_documentacion`, `estado_servicios`, `buscar_servidores`. Las de incidentes, certificados y deploys se agregan cuando existan esos CU.
- Ninguna tool puede llamar métodos que modifiquen datos. Hay un test que lo verifica.
- Cada respuesta incluye `sources: [{type: "document"|"server"|"service", id, title}]`, registradas por las tools (no por el LLM).
- Si no hay contexto suficiente, responder que no hay información; no inventar.
- Metadatos de cada chunk en Chroma: `document_id`, `title`, `section`, `version`, `tags`.
- Reindexado incremental: borrar los chunks del `document_id` y volver a insertarlos.

## System prompt (base)

> Sos OPSBot, un ingeniero SRE de guardia del equipo. Respondés en español, de forma concisa.
> Priorizás los procedimientos documentados en los runbooks. Siempre citás la fuente.
> Si una sugerencia no proviene de un runbook, lo aclarás. Antes de sugerir una acción destructiva
> (borrar, reiniciar, restaurar) lo advertís explícitamente. Si no tenés información suficiente, lo decís.

## Tests

- pytest + pytest-asyncio. Tests del splitter, de la sincronización (Chroma en memoria), de las tools con el backend
  mockeado (respx), del endpoint `/chat` con un LLM simulado y un test que verifica que solo se hacen GET (CA-4).

## Verificación (loop de autocorrección)

```bash
ruff check . && ruff format --check .
mypy .
pytest
```
