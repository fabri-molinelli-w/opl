"""Acceso a ChromaDB: una colección con los fragmentos de la documentación de OPL."""

import uuid
from dataclasses import dataclass
from typing import Any

import chromadb

from app.rag.embeddings import Embedder
from app.rag.splitter import split_markdown

COLLECTION = "documents"


@dataclass(frozen=True)
class Hit:
    document_id: str
    title: str
    section: str
    text: str
    distance: float


class DocumentStore:
    def __init__(self, client: Any, embedder: Embedder, collection: str = COLLECTION) -> None:
        # Any: chromadb no exporta tipos públicos para el cliente.
        self._embedder = embedder
        self._collection = client.get_or_create_collection(
            name=collection, metadata={"hnsw:space": "cosine"}, embedding_function=None
        )

    @classmethod
    def persistent(cls, path: str, embedder: Embedder) -> "DocumentStore":
        return cls(chromadb.PersistentClient(path=path), embedder)

    @classmethod
    def in_memory(cls, embedder: Embedder) -> "DocumentStore":
        # Los clientes efímeros comparten estado en el proceso: cada store usa su propia colección.
        return cls(chromadb.EphemeralClient(), embedder, collection=f"test-{uuid.uuid4().hex}")

    def indexed_versions(self) -> dict[str, int]:
        result = self._collection.get(include=["metadatas"])
        versions: dict[str, int] = {}
        for metadata in result["metadatas"] or []:
            versions[str(metadata["document_id"])] = int(metadata["version"])
        return versions

    def replace_document(self, document_id: str, title: str, version: int, tags: list[str], content: str) -> int:
        """Reindexado incremental: borra los fragmentos del documento y los vuelve a insertar."""
        self.delete_document(document_id)
        chunks = split_markdown(title, content)
        texts = [chunk.text for chunk in chunks]
        self._collection.add(
            ids=[f"{document_id}:{i}" for i in range(len(chunks))],
            documents=texts,
            embeddings=self._embedder.embed_documents(texts),
            metadatas=[
                {
                    "document_id": document_id,
                    "title": title,
                    "section": chunk.section,
                    "version": version,
                    "tags": ",".join(tags),
                }
                for chunk in chunks
            ],
        )
        return len(chunks)

    def delete_document(self, document_id: str) -> None:
        self._collection.delete(where={"document_id": document_id})

    def search(self, query: str, k: int) -> list[Hit]:
        if self._collection.count() == 0:
            return []
        result = self._collection.query(
            query_embeddings=[self._embedder.embed_query(query)],
            n_results=min(k, self._collection.count()),
            include=["documents", "metadatas", "distances"],
        )
        hits: list[Hit] = []
        for text, metadata, distance in zip(
            result["documents"][0], result["metadatas"][0], result["distances"][0], strict=True
        ):
            hits.append(
                Hit(
                    document_id=str(metadata["document_id"]),
                    title=str(metadata["title"]),
                    section=str(metadata["section"]),
                    text=str(text),
                    distance=float(distance),
                )
            )
        return hits
