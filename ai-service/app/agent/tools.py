"""Tools de OPSBot. Todas son de SOLO LECTURA (CU-09 RN-5): solo consultan Chroma o hacen GET al backend."""

from dataclasses import dataclass, field
from typing import Literal

from langchain_core.tools import BaseTool, StructuredTool

from app.backend_client import BackendClient, BackendError
from app.rag.store import DocumentStore
from app.schemas import Source, SourceType

Environment = Literal["prod", "staging", "dev"]
ServiceStatus = Literal["UP", "DOWN", "DEGRADED", "UNKNOWN"]
RELEVANCE_GAP = 0.15


@dataclass
class SourceCollector:
    """Junta las fuentes que realmente usaron las tools, para devolverlas con la respuesta."""

    items: list[Source] = field(default_factory=list)

    def add(self, type_: SourceType, id_: str, title: str) -> None:
        if not any(s.type == type_ and s.id == id_ for s in self.items):
            self.items.append(Source(type=type_, id=id_, title=title))


def build_tools(client: BackendClient, store: DocumentStore, sources: SourceCollector, top_k: int) -> list[BaseTool]:
    async def buscar_documentacion(consulta: str) -> str:
        """Busca en la documentación del equipo (procedimientos, runbooks, guías, configuraciones).

        Usala para cualquier pregunta de "cómo hago...", pasos, comandos o procedimientos.
        """
        hits = store.search(consulta, top_k)
        if not hits:
            return "No hay documentación relacionada con esa consulta."
        # Solo se citan los fragmentos cercanos al mejor resultado (evita fuentes poco relacionadas).
        hits = [hit for hit in hits if hit.distance <= hits[0].distance + RELEVANCE_GAP]
        parts: list[str] = []
        for hit in hits:
            sources.add("document", hit.document_id, hit.title)
            section = f" / {hit.section}" if hit.section else ""
            parts.append(f"[Documento: {hit.title}{section}]\n{hit.text}")
        return "\n\n---\n\n".join(parts)

    async def estado_servicios(environment: Environment | None = None, status: ServiceStatus | None = None) -> str:
        """Lista los servicios del inventario con su estado actual (UP, DOWN, DEGRADED, UNKNOWN).

        Filtros opcionales: environment (prod, staging, dev) y status. Para "servicios caídos" usá status=DOWN.
        """
        params: dict[str, str | int] = {"pageSize": 100}
        if environment:
            params["environment"] = environment
        if status:
            params["status"] = status
        try:
            services = await client.list_services(params)
            servers = await client.list_servers({"pageSize": 100})
        except BackendError as exc:
            return f"No se pudo consultar el estado de los servicios ({exc})."
        hostnames = {str(s["id"]): str(s["hostname"]) for s in servers["items"]}
        items = services["items"]
        if not items:
            return "No hay servicios que coincidan con esos filtros."
        lines: list[str] = []
        for svc in items:
            server_id = str(svc["serverId"])
            hostname = hostnames.get(server_id, server_id)
            sources.add("server", server_id, hostname)
            lines.append(
                f"- {svc['name']} en {hostname}: {svc['currentStatus']} "
                f"(tipo {svc['type']}, puerto {svc['port']}, chequeo {svc['checkType']} {svc['checkTarget']})"
            )
        return f"{len(items)} servicio(s):\n" + "\n".join(lines)

    async def buscar_servidores(texto: str | None = None, environment: Environment | None = None) -> str:
        """Busca servidores del inventario por hostname, IP o responsable, con filtro opcional por environment."""
        params: dict[str, str | int] = {"pageSize": 50}
        if texto:
            params["q"] = texto
        if environment:
            params["environment"] = environment
        try:
            servers = await client.list_servers(params)
        except BackendError as exc:
            return f"No se pudo consultar el inventario ({exc})."
        items = servers["items"]
        if not items:
            return "No hay servidores que coincidan."
        lines: list[str] = []
        for srv in items:
            sources.add("server", str(srv["id"]), str(srv["hostname"]))
            lines.append(
                f"- {srv['hostname']} ({srv['ip']}), {srv['environment']}, SO {srv['os'] or 'sin dato'}, "
                f"responsable {srv['owner'] or 'sin dato'}, etiquetas: {', '.join(srv['tags']) or 'ninguna'}"
            )
        return f"{len(items)} servidor(es):\n" + "\n".join(lines)

    async def incidentes_activos() -> str:
        """Lista los incidentes abiertos o en curso, con severidad, servicio afectado y responsable."""
        try:
            data = await client.list_incidents({"active": "true", "pageSize": 50})
        except BackendError as exc:
            return f"No se pudo consultar los incidentes ({exc})."
        items = data["items"]
        if not items:
            return "No hay incidentes abiertos ni en curso."
        lines: list[str] = []
        for inc in items:
            sources.add("incident", str(inc["id"]), str(inc["title"]))
            origin = "automático" if inc["autoCreated"] else "manual"
            service = inc["serviceName"] or "ninguno"
            assignee = inc["assigneeName"] or "sin asignar"
            lines.append(
                f"- [{inc['severity']}] {inc['title']} — estado {inc['status']}, servicio {service}, "
                f"responsable {assignee}, abierto {inc['openedAt']} ({origin})"
            )
        return f"{len(items)} incidente(s) activo(s):\n" + "\n".join(lines)

    return [
        StructuredTool.from_function(coroutine=buscar_documentacion, name="buscar_documentacion"),
        StructuredTool.from_function(coroutine=incidentes_activos, name="incidentes_activos"),
        StructuredTool.from_function(coroutine=estado_servicios, name="estado_servicios"),
        StructuredTool.from_function(coroutine=buscar_servidores, name="buscar_servidores"),
    ]
