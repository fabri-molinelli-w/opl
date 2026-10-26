"""FastAPI de OPSBot. Servicio interno: solo lo llama el backend Go, nunca el navegador."""

import hmac
import logging
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager
from dataclasses import dataclass
from typing import Annotated

import httpx
from fastapi import Depends, FastAPI, Header, HTTPException, Request
from langchain_core.language_models import BaseChatModel

from app.agent.agent import build_llm, run_agent
from app.agent.memory import SessionMemory
from app.agent.tools import SourceCollector, build_tools
from app.backend_client import BackendClient, BackendError
from app.config import Settings, get_settings
from app.rag.embeddings import build_embedder
from app.rag.store import DocumentStore
from app.rag.sync import sync_documents
from app.schemas import ChatRequest, ChatResponse

log = logging.getLogger("opsbot")


@dataclass
class AppState:
    settings: Settings
    http: httpx.AsyncClient
    store: DocumentStore
    llm: BaseChatModel
    memory: SessionMemory


def create_app(state: AppState | None = None) -> FastAPI:
    """state permite inyectar dependencias en los tests (LLM simulado, Chroma en memoria)."""

    @asynccontextmanager
    async def lifespan(app: FastAPI) -> AsyncIterator[None]:
        if state is not None:
            app.state.opsbot = state
            yield
            return
        settings = get_settings()
        async with httpx.AsyncClient() as http:
            app.state.opsbot = AppState(
                settings=settings,
                http=http,
                store=DocumentStore.persistent(settings.chroma_path, build_embedder(settings.embeddings_model)),
                llm=build_llm(settings),
                memory=SessionMemory(settings.memory_turns),
            )
            yield

    app = FastAPI(title="OPSBot", lifespan=lifespan, docs_url=None, redoc_url=None)

    def get_state(request: Request) -> AppState:
        opsbot: AppState = request.app.state.opsbot
        return opsbot

    def require_internal_token(request: Request, x_internal_token: Annotated[str | None, Header()] = None) -> None:
        expected = get_state(request).settings.ai_internal_token.get_secret_value()
        if x_internal_token is None or not hmac.compare_digest(x_internal_token, expected):
            raise HTTPException(status_code=401, detail="token interno inválido")

    @app.get("/health")
    async def health() -> dict[str, str]:
        return {"status": "ok"}

    @app.post("/chat", dependencies=[Depends(require_internal_token)], response_model_by_alias=True)
    async def chat(
        body: ChatRequest,
        request: Request,
        authorization: Annotated[str | None, Header()] = None,
        x_user_id: Annotated[str | None, Header()] = None,
    ) -> ChatResponse:
        if authorization is None or not authorization.startswith("Bearer ") or not x_user_id:
            raise HTTPException(status_code=401, detail="falta el usuario")
        opsbot = get_state(request)
        client = BackendClient(opsbot.http, opsbot.settings.backend_url, authorization.removeprefix("Bearer "))

        try:
            await sync_documents(client, opsbot.store)
        except BackendError as exc:
            # Sin sincronizar igual se puede responder con lo ya indexado.
            log.warning("document sync failed: %s", exc)

        sources = SourceCollector()
        tools = build_tools(client, opsbot.store, sources, opsbot.settings.rag_top_k)
        key = SessionMemory.key(x_user_id, body.session_id)
        try:
            answer = await run_agent(opsbot.llm, tools, opsbot.memory.history(key), body.message)
        except Exception as exc:  # el proveedor del LLM puede fallar de muchas formas
            log.exception("llm call failed")
            raise HTTPException(status_code=502, detail="el modelo de lenguaje no respondió") from exc
        opsbot.memory.append_turn(key, body.message, answer)
        return ChatResponse(answer=answer, sources=sources.items)

    return app


app = create_app()
