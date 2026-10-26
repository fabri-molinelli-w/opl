from typing import Any

import httpx
import pytest
import respx
from fastapi.testclient import TestClient
from langchain_core.messages import AIMessage, HumanMessage

from app.agent.memory import SessionMemory
from app.config import get_settings
from app.main import AppState, create_app
from app.rag.embeddings import HashEmbedder
from app.rag.store import DocumentStore
from tests.conftest import BACKEND, ScriptedChatModel

HEADERS = {"X-Internal-Token": "test-internal-token", "Authorization": "Bearer jwt", "X-User-Id": "user-1"}


def make_client(llm: ScriptedChatModel) -> tuple[TestClient, AppState]:
    settings = get_settings().model_copy(update={"backend_url": BACKEND})
    state = AppState(
        settings=settings,
        http=httpx.AsyncClient(),
        store=DocumentStore.in_memory(HashEmbedder()),
        llm=llm,
        memory=SessionMemory(turns=10),
    )
    return TestClient(create_app(state=state)), state


@pytest.fixture
def backend(documents_page: dict[str, Any], documents: dict[str, dict[str, Any]]) -> respx.MockRouter:
    router = respx.mock(assert_all_called=False)
    router.get(f"{BACKEND}/api/documents").mock(return_value=httpx.Response(200, json=documents_page))
    for doc_id, doc in documents.items():
        router.get(f"{BACKEND}/api/documents/{doc_id}").mock(return_value=httpx.Response(200, json=doc))
    return router


def test_chat_requires_internal_token() -> None:
    client, _ = make_client(ScriptedChatModel(script=[]))
    with client:
        response = client.post(
            "/chat",
            json={"sessionId": "s1", "message": "hola"},
            headers={**HEADERS, "X-Internal-Token": "otro"},
        )
    assert response.status_code == 401


def test_chat_uses_rag_and_returns_sources(backend: respx.MockRouter) -> None:
    llm = ScriptedChatModel(
        script=[
            AIMessage(
                content="",
                tool_calls=[
                    {
                        "name": "buscar_documentacion",
                        "args": {"consulta": "reload nginx"},
                        "id": "call-1",
                        "type": "tool_call",
                    }
                ],
            ),
            AIMessage(content="Según **Reiniciar servicio Nginx**: `sudo nginx -t` y después `reload`."),
        ]
    )
    client, state = make_client(llm)
    with backend, client:
        response = client.post("/chat", json={"sessionId": "s1", "message": "¿Cómo recargo nginx?"}, headers=HEADERS)

    assert response.status_code == 200, response.text
    body = response.json()
    assert "Reiniciar servicio Nginx" in body["answer"]
    assert {"type": "document", "id": "doc-nginx", "title": "Reiniciar servicio Nginx"} in body["sources"]
    # El resultado de la tool (fragmento del documento) llegó al modelo.
    tool_result = llm.received[1][-1]
    assert "systemctl reload nginx" in str(tool_result.content)
    # Memoria por usuario + sesión (RN-3).
    history = state.memory.history(SessionMemory.key("user-1", "s1"))
    assert [type(m) for m in history] == [HumanMessage, AIMessage]
    assert state.memory.history(SessionMemory.key("user-2", "s1")) == []


def test_chat_without_tools_has_no_sources(backend: respx.MockRouter) -> None:
    llm = ScriptedChatModel(script=[AIMessage(content="No tengo información sobre eso en OPL.")])
    client, _ = make_client(llm)
    with backend, client:
        response = client.post("/chat", json={"sessionId": "s2", "message": "¿Quién ganó el mundial?"}, headers=HEADERS)
    assert response.status_code == 200
    assert response.json() == {"answer": "No tengo información sobre eso en OPL.", "sources": []}


def test_chat_llm_failure_returns_502(backend: respx.MockRouter) -> None:
    client, _ = make_client(ScriptedChatModel(script=[]))  # script vacío → el LLM falla
    with backend, client:
        response = client.post("/chat", json={"sessionId": "s3", "message": "hola"}, headers=HEADERS)
    assert response.status_code == 502
