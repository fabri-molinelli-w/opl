"""CA-4: ninguna tool de OPSBot puede modificar datos (solo GET)."""

import ast
import inspect
from pathlib import Path
from typing import Any

import httpx
import respx

import app.backend_client as backend_client_module
from app.agent.tools import SourceCollector, build_tools
from app.backend_client import BackendClient
from app.rag.embeddings import HashEmbedder
from app.rag.store import DocumentStore
from tests.conftest import BACKEND

WRITE_METHODS = {"post", "put", "patch", "delete", "request", "stream", "send"}


def test_backend_client_source_only_uses_get() -> None:
    tree = ast.parse(Path(inspect.getfile(backend_client_module)).read_text())
    called = {
        node.func.attr for node in ast.walk(tree) if isinstance(node, ast.Call) and isinstance(node.func, ast.Attribute)
    }
    assert not called & WRITE_METHODS, f"el cliente usa métodos de escritura: {called & WRITE_METHODS}"


def test_backend_client_public_api_has_no_write_operations() -> None:
    public = [name for name in dir(BackendClient) if not name.startswith("_")]
    assert all(name.startswith(("list_", "get_")) for name in public), public


async def test_every_tool_only_sends_get_requests() -> None:
    servers = {
        "items": [
            {
                "id": "srv-1",
                "hostname": "web-01.example.com",
                "ip": "192.0.2.10",
                "environment": "prod",
                "os": "debian-12",
                "owner": "Plataforma",
                "tags": ["web"],
            }
        ]
    }
    services: dict[str, Any] = {
        "items": [
            {
                "serverId": "srv-1",
                "name": "Portal",
                "currentStatus": "DOWN",
                "type": "web",
                "port": 443,
                "checkType": "http",
                "checkTarget": "https://web-01.example.com/health",
            }
        ]
    }
    store = DocumentStore.in_memory(HashEmbedder())
    store.replace_document("doc-1", "Reiniciar Nginx", 1, [], "sudo systemctl reload nginx")

    async with httpx.AsyncClient() as http:
        with respx.mock(assert_all_called=False) as router:
            router.get(f"{BACKEND}/api/servers").mock(return_value=httpx.Response(200, json=servers))
            router.get(f"{BACKEND}/api/services").mock(return_value=httpx.Response(200, json=services))
            sources = SourceCollector()
            tools = build_tools(BackendClient(http, BACKEND, "jwt"), store, sources, top_k=3)
            args: dict[str, dict[str, str]] = {
                "buscar_documentacion": {"consulta": "nginx"},
                "estado_servicios": {"environment": "prod", "status": "DOWN"},
                "buscar_servidores": {"texto": "web"},
            }
            assert {t.name for t in tools} == set(args)
            for tool in tools:
                await tool.ainvoke(args[tool.name])

            assert router.calls
            assert {call.request.method for call in router.calls} == {"GET"}

    # CA-1: las tools registran de dónde salió la información.
    assert {(s.type, s.id) for s in sources.items} == {("document", "doc-1"), ("server", "srv-1")}
