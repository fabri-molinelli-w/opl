"""Cliente HTTP hacia el backend Go.

Es deliberadamente de SOLO LECTURA: la única operación HTTP que expone es GET (CU-09 RN-5).
Todas las llamadas viajan con el JWT del usuario, así el backend aplica sus permisos.
"""

from typing import Any, cast

import httpx

JSON = dict[str, Any]  # Any: JSON del backend, se valida al leer cada campo.


class BackendError(Exception):
    """El backend respondió con un error o no se pudo contactar."""

    def __init__(self, status: int, message: str) -> None:
        super().__init__(message)
        self.status = status


class BackendClient:
    def __init__(self, http: httpx.AsyncClient, base_url: str, user_token: str) -> None:
        self._http = http
        self._base_url = base_url.rstrip("/")
        self._headers = {"Authorization": f"Bearer {user_token}", "Accept": "application/json"}

    async def _get(self, path: str, params: dict[str, str | int] | None = None) -> Any:
        try:
            response = await self._http.get(
                f"{self._base_url}{path}", params=params, headers=self._headers, timeout=15.0
            )
        except httpx.HTTPError as exc:
            raise BackendError(503, "no se pudo contactar al backend") from exc
        if response.status_code >= 400:
            raise BackendError(response.status_code, f"el backend respondió {response.status_code}")
        return response.json()

    async def list_documents(self, page: int, page_size: int = 100) -> JSON:
        return cast(JSON, await self._get("/api/documents", {"page": page, "pageSize": page_size}))

    async def get_document(self, document_id: str) -> JSON:
        return cast(JSON, await self._get(f"/api/documents/{document_id}"))

    async def list_services(self, params: dict[str, str | int]) -> JSON:
        return cast(JSON, await self._get("/api/services", params))

    async def list_servers(self, params: dict[str, str | int]) -> JSON:
        return cast(JSON, await self._get("/api/servers", params))

    async def list_incidents(self, params: dict[str, str | int]) -> JSON:
        return cast(JSON, await self._get("/api/incidents", params))
