import os
from typing import Any

import pytest
from langchain_core.callbacks import CallbackManagerForLLMRun
from langchain_core.language_models import BaseChatModel
from langchain_core.messages import AIMessage, BaseMessage
from langchain_core.outputs import ChatGeneration, ChatResult

# Valores ficticios para que Settings se pueda construir en los tests.
os.environ.setdefault("AI_INTERNAL_TOKEN", "test-internal-token")
os.environ.setdefault("LLM_BASE_URL", "http://llm.invalid/v1")
os.environ.setdefault("LLM_API_KEY", "test-key")
os.environ.setdefault("LLM_MODEL", "fake-model")

BACKEND = "http://backend.test"


class ScriptedChatModel(BaseChatModel):
    """LLM simulado: devuelve en orden los mensajes del guion y registra lo que recibió."""

    script: list[AIMessage]
    received: list[list[BaseMessage]] = []

    @property
    def _llm_type(self) -> str:
        return "scripted"

    def bind_tools(self, tools: Any, **kwargs: Any) -> "ScriptedChatModel":  # noqa: ARG002
        return self

    def _generate(
        self,
        messages: list[BaseMessage],
        stop: list[str] | None = None,
        run_manager: CallbackManagerForLLMRun | None = None,
        **kwargs: Any,
    ) -> ChatResult:
        self.received.append(list(messages))
        message = self.script.pop(0)
        return ChatResult(generations=[ChatGeneration(message=message)])


@pytest.fixture
def documents_page() -> dict[str, Any]:
    return {
        "items": [
            {"id": "doc-vpn", "title": "Configurar VPN en notebook", "version": 1},
            {"id": "doc-nginx", "title": "Reiniciar servicio Nginx", "version": 2},
        ],
        "total": 2,
        "page": 1,
        "pageSize": 100,
    }


@pytest.fixture
def documents() -> dict[str, dict[str, Any]]:
    return {
        "doc-vpn": {
            "id": "doc-vpn",
            "title": "Configurar VPN en notebook",
            "version": 1,
            "tags": ["vpn"],
            "contentMd": "## Pasos\n\n1. Instalar el cliente WireGuard\n2. Importar el perfil de la VPN",
        },
        "doc-nginx": {
            "id": "doc-nginx",
            "title": "Reiniciar servicio Nginx",
            "version": 2,
            "tags": ["nginx"],
            "contentMd": "## Pasos\n\n```bash\nsudo nginx -t\nsudo systemctl reload nginx\n```",
        },
    }
