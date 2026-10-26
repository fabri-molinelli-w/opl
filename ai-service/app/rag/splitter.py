"""Chunking de Markdown por encabezados, con tamaño máximo y solapamiento."""

import re
from dataclasses import dataclass

_HEADING = re.compile(r"^(#{1,3})\s+(.+?)\s*$")


@dataclass(frozen=True)
class Chunk:
    section: str
    text: str


def split_markdown(title: str, content: str, max_chars: int = 1200, overlap: int = 150) -> list[Chunk]:
    """Divide el documento en secciones por encabezado (#, ##, ###).

    Cada chunk empieza con el título del documento y la sección, para que el embedding y el
    LLM tengan contexto aunque el fragmento sea chico. Las secciones largas se cortan con
    solapamiento. Los encabezados dentro de bloques de código no cuentan.
    """
    sections: list[tuple[str, list[str]]] = [("", [])]
    in_code = False
    for line in content.splitlines():
        if line.lstrip().startswith("```"):
            in_code = not in_code
        match = None if in_code else _HEADING.match(line)
        if match:
            sections.append((match.group(2), []))
        else:
            sections[-1][1].append(line)

    chunks: list[Chunk] = []
    for section, lines in sections:
        body = "\n".join(lines).strip()
        if not body:
            continue
        prefix = f"{title}\n{section}\n" if section else f"{title}\n"
        for piece in _window(body, max_chars, overlap):
            chunks.append(Chunk(section=section, text=prefix + piece))
    if not chunks:
        chunks.append(Chunk(section="", text=title))
    return chunks


def _window(text: str, size: int, overlap: int) -> list[str]:
    if len(text) <= size:
        return [text]
    pieces: list[str] = []
    start = 0
    while start < len(text):
        end = min(start + size, len(text))
        pieces.append(text[start:end])
        if end == len(text):
            break
        start = end - overlap
    return pieces
