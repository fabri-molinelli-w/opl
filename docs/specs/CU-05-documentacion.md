# CU-05 · Base de documentación (runbooks y procedimientos)

| Campo | Valor |
|---|---|
| Actor principal | Soporte IT y SysAdmin |
| Prioridad | **Alta — foco principal de la primera etapa** |
| Depende de | CU-01; CU-02 para vincular documentos a servidores/servicios (opcional) |

## Descripción
Repositorio central de la documentación del área: procedimientos para resolver problemas, guías de instalación, configuraciones, notas de referencia. Reemplaza los Word sueltos y los procedimientos
"en la cabeza de alguien". Más adelante alimenta a OPLBot (CU-09).

## Conceptos
- **Documento**: título + contenido en Markdown + categoría + etiquetas.
- **Categoría** (fija): `procedimiento`, `runbook`, `guia`, `configuracion`, `referencia`.
- **Etiquetas**: libres (ej. `impresoras`, `vpn`, `active-directory`, `backup`).
- **Vínculos**: un documento puede asociarse a uno o más servidores y/o servicios del inventario.
- **Versión**: cada vez que se guarda un cambio se conserva la versión anterior.

## Reglas de negocio
- RN-1 Título obligatorio (3-150 caracteres) y único entre documentos no eliminados (409).
- RN-2 Contenido Markdown obligatorio.
- RN-3 `soporte`, `sysadmin` y `admin` pueden crear y editar. Solo `sysadmin` y `admin` pueden eliminar (baja lógica).
- RN-4 Al editar, se guarda una fila en `document_versions` con el contenido anterior, autor y fecha, en la **misma transacción**.
- RN-5 Se registra `created_by` y `updated_by`.
- RN-6 Búsqueda por texto en título, contenido y etiquetas, más filtros por categoría, etiqueta, servidor y servicio. MVP con `ILIKE`; si queda lento, se pasa a full-text search de PostgreSQL (`tsvector` en español).
- RN-7 Listado paginado y ordenable por título o por última actualización.
- RN-9 Bloqueo optimista: el PATCH puede incluir la `version` editada; si no coincide con la actual → 409 (`details[].field = "version"`).
- RN-10 El Markdown se renderiza sin HTML crudo (sin `rehype-raw` ni `dangerouslySetInnerHTML`). Los links `javascript:` no se ejecutan.
- RN-11 Contenido de hasta 200 KB; hasta 20 etiquetas de hasta 40 caracteres, normalizadas a minúsculas y sin duplicados.
- RN-8 (Fase OPLBot) Al crear/editar/eliminar se avisa al ai-service para reindexar. Si no responde, se guarda igual con `index_status = 'pending'`.

## Modelo de datos
- `documents(id, title, content_md text, category text check in (...), tags text[], version int default 1, created_by fk users, updated_by fk users, created_at, updated_at, deleted_at)` — migración `00005`
- `document_versions(id, document_id fk, version int, title, content_md, category, tags, edited_by fk users, created_at)` — único `(document_id, version)`. Snapshot completo; la versión 1 se guarda al crear.
- *(Postergado)* `document_links(document_id fk, server_id fk null, service_id fk null)` y `index_status`: ver "Fuera de alcance".

## Endpoints
| Método | Ruta | Rol mínimo |
|---|---|---|
| GET | /api/documents?q=&category=&tag=&sort=title\|updatedAt&order=asc\|desc&page=&pageSize= | soporte |
| POST | /api/documents | soporte |
| GET | /api/documents/{id} | soporte |
| PATCH | /api/documents/{id} | soporte |
| DELETE | /api/documents/{id} | sysadmin |
| GET | /api/documents/{id}/versions | soporte (solo metadatos) |
| GET | /api/documents/{id}/versions/{version} | soporte |
| GET | /api/documents/tags | soporte (lista de etiquetas en uso, para autocompletar) |

## UI
- **Listado**: buscador grande arriba, filtros por categoría y etiqueta, tarjetas o tabla con título, categoría, etiquetas, última edición y autor.
- **Vista de documento**: Markdown renderizado (títulos, listas, bloques de código con botón "copiar"), etiquetas, servidores/servicios vinculados, historial de versiones.
- **Editor**: Markdown con vista previa lado a lado; selector de categoría, etiquetas con autocompletado y vínculos al inventario.
- En la ficha de un servidor o servicio (CU-02) se listan sus documentos vinculados.

## Datos semilla (ficticios)
"Reiniciar servicio Nginx", "Alta de usuario en Active Directory", "Configurar VPN en notebook", "Restaurar backup de PostgreSQL", "Renovar certificado TLS", "Impresora no imprime: checklist".

## Criterios de aceptación
- [ ] CA-1 Un usuario `soporte` crea un documento con categoría y etiquetas y lo encuentra por búsqueda.
- [ ] CA-2 Editar un documento genera una versión nueva y la anterior se puede consultar.
- [ ] CA-3 Título repetido → 409.
- [ ] CA-4 Un usuario `soporte` que intenta eliminar recibe 403.
- [ ] CA-5 Los bloques de código del Markdown se muestran con botón de copiar.
- [ ] CA-6 Guardar con una versión desactualizada → 409 con mensaje claro y sin perder lo escrito.
- [ ] CA-7 Un documento con `<script>` o un link `javascript:` no ejecuta código al visualizarse.

## Fuera de alcance (por ahora)
Vínculos con servidores/servicios (`document_links`) y su listado en la ficha del servidor (se pospone por la fecha de entrega), adjuntos (PDF, imágenes), comentarios, edición colaborativa en tiempo real, permisos por documento.
