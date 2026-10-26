"""Embeddings locales: el contenido nunca sale del servidor para vectorizarse (ADR-005/ADR-022)."""

import hashlib
import math
from typing import Protocol


class Embedder(Protocol):
    def embed_documents(self, texts: list[str]) -> list[list[float]]: ...

    def embed_query(self, text: str) -> list[float]: ...


class FastEmbedEmbedder:
    """Modelo multilingüe vía fastembed (ONNX, sin PyTorch). Se descarga una vez y queda en caché."""

    def __init__(self, model_name: str) -> None:
        from fastembed import TextEmbedding  # import diferido: carga el runtime ONNX

        self._model = TextEmbedding(model_name=model_name)

    def embed_documents(self, texts: list[str]) -> list[list[float]]:
        return [vector.tolist() for vector in self._model.embed(texts)]

    def embed_query(self, text: str) -> list[float]:
        return self.embed_documents([text])[0]


class HashEmbedder:
    """Embedder determinístico para tests (bolsa de palabras con hashing). No usar en producción."""

    def __init__(self, dimensions: int = 256) -> None:
        self._dimensions = dimensions

    def embed_documents(self, texts: list[str]) -> list[list[float]]:
        return [self.embed_query(text) for text in texts]

    def embed_query(self, text: str) -> list[float]:
        vector = [0.0] * self._dimensions
        for word in text.lower().split():
            token = word.strip(".,:;!?¿¡()[]`*#\"'")
            if not token:
                continue
            digest = hashlib.sha256(token.encode()).digest()
            vector[int.from_bytes(digest[:4], "big") % self._dimensions] += 1.0
        norm = math.sqrt(sum(v * v for v in vector)) or 1.0
        return [v / norm for v in vector]


def build_embedder(model_name: str) -> Embedder:
    """ "hash" es solo para desarrollo sin internet (pruebas de punta a punta); no usar en producción."""
    if model_name == "hash":
        return HashEmbedder()
    return FastEmbedEmbedder(model_name)
