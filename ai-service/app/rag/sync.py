"""Sincronización de la documentación con ChromaDB.

En lugar de que el backend empuje cambios, OPSBot compara antes de responder la versión de cada
documento (GET /api/documents, sin contenido) con la indexada, y solo descarga los que cambiaron.
Así el ai-service sigue siendo de solo lectura y no hace falta un endpoint de escritura interno.
"""

import asyncio
import logging
from dataclasses import dataclass

from app.backend_client import BackendClient
from app.rag.store import DocumentStore

log = logging.getLogger(__name__)
_lock = asyncio.Lock()


@dataclass(frozen=True)
class SyncResult:
    indexed: int
    deleted: int


async def sync_documents(client: BackendClient, store: DocumentStore) -> SyncResult:
    async with _lock:
        current: dict[str, int] = {}
        page = 1
        while True:
            data = await client.list_documents(page)
            for item in data["items"]:
                current[str(item["id"])] = int(item["version"])
            total = int(data["total"])
            if page * int(data["pageSize"]) >= total:
                break
            page += 1

        indexed = store.indexed_versions()
        changed = [doc_id for doc_id, version in current.items() if indexed.get(doc_id) != version]
        removed = [doc_id for doc_id in indexed if doc_id not in current]

        for doc_id in changed:
            doc = await client.get_document(doc_id)
            store.replace_document(
                document_id=doc_id,
                title=str(doc["title"]),
                version=int(doc["version"]),
                tags=[str(tag) for tag in doc["tags"]],
                content=str(doc["contentMd"]),
            )
        for doc_id in removed:
            store.delete_document(doc_id)

        if changed or removed:
            log.info("documents synced", extra={"indexed": len(changed), "deleted": len(removed)})
        return SyncResult(indexed=len(changed), deleted=len(removed))
