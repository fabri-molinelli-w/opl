from typing import Any

import httpx
import respx

from app.backend_client import BackendClient
from app.rag.embeddings import HashEmbedder
from app.rag.splitter import split_markdown
from app.rag.store import DocumentStore
from app.rag.sync import sync_documents
from tests.conftest import BACKEND


def test_split_by_headings_ignores_headings_inside_code() -> None:
    content = "Intro\n\n## Pasos\n\n```bash\n# esto es un comentario\nls\n```\n\n## Verificación\n\nok"
    chunks = split_markdown("Doc", content)
    assert [c.section for c in chunks] == ["", "Pasos", "Verificación"]
    assert "# esto es un comentario" in chunks[1].text
    assert all(c.text.startswith("Doc\n") for c in chunks)


def test_split_long_section_with_overlap() -> None:
    chunks = split_markdown("Doc", "## A\n\n" + "x" * 3000, max_chars=1000, overlap=100)
    assert len(chunks) == 4
    assert all(c.section == "A" for c in chunks)


def _mock_backend(router: respx.Router, page: dict[str, Any], docs: dict[str, dict[str, Any]]) -> None:
    router.get(f"{BACKEND}/api/documents").mock(return_value=httpx.Response(200, json=page))
    for doc_id, doc in docs.items():
        router.get(f"{BACKEND}/api/documents/{doc_id}").mock(return_value=httpx.Response(200, json=doc))


async def test_sync_indexes_only_changes_and_deletes_removed(
    documents_page: dict[str, Any], documents: dict[str, dict[str, Any]]
) -> None:
    store = DocumentStore.in_memory(HashEmbedder())
    store.replace_document("doc-old", "Documento borrado", 1, [], "contenido viejo")

    async with httpx.AsyncClient() as http:
        client = BackendClient(http, BACKEND, "jwt")
        with respx.mock(assert_all_called=False) as router:
            _mock_backend(router, documents_page, documents)
            first = await sync_documents(client, store)
            assert (first.indexed, first.deleted) == (2, 1)

            second = await sync_documents(client, store)
            assert (second.indexed, second.deleted) == (0, 0)

            # Una nueva versión de un documento se reindexa sola.
            documents_page["items"][0]["version"] = 2
            documents["doc-vpn"]["version"] = 2
            documents["doc-vpn"]["contentMd"] = "## Pasos\n\nUsar el cliente OpenVPN"
            _mock_backend(router, documents_page, documents)
            third = await sync_documents(client, store)
            assert (third.indexed, third.deleted) == (1, 0)

    assert store.indexed_versions() == {"doc-vpn": 2, "doc-nginx": 2}
    hits = store.search("cliente openvpn", 1)
    assert hits[0].document_id == "doc-vpn"
    assert "OpenVPN" in hits[0].text


async def test_sync_sends_user_jwt(documents_page: dict[str, Any], documents: dict[str, dict[str, Any]]) -> None:
    store = DocumentStore.in_memory(HashEmbedder())
    async with httpx.AsyncClient() as http:
        with respx.mock(assert_all_called=False) as router:
            _mock_backend(router, documents_page, documents)
            await sync_documents(BackendClient(http, BACKEND, "jwt-del-usuario"), store)
            assert router.calls
            for call in router.calls:
                assert call.request.headers["Authorization"] == "Bearer jwt-del-usuario"
