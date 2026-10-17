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
- RN-8 (Fase OPLBot) Al crear/editar/eliminar se avisa al ai-service para reindexar. Si no responde, se guarda igual con `index_status = 'pending'`.

## Modelo de datos
- `documents(id, title, content_md text, category text check in (...), tags text[], index_status text default 'pending', created_by fk users, updated_by fk users, created_at, updated_at, deleted_at)`
- `document_versions(id, document_id fk, version int, title, content_md, edited_by fk users, created_at)` — único `(document_id, version)`
- `document_links(document_id fk, server_id fk null, service_id fk null)` — al menos uno de los dos no nulo

## Endpoints
| Método | Ruta | Rol mínimo |
|---|---|---|
| GET | /api/documents?q=&category=&tag=&server_id=&service_id=&sort=&page= | soporte |
| POST | /api/documents | soporte |
| GET | /api/documents/{id} | soporte |
| PATCH | /api/documents/{id} | soporte |
| DELETE | /api/documents/{id} | sysadmin |
| GET | /api/documents/{id}/versions | soporte |
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

## Fuera de alcance (por ahora)
Adjuntos (PDF, imágenes), comentarios, edición colaborativa en tiempo real, permisos por documento.
