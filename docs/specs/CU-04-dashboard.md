# CU-04 · Dashboard de estado de la infraestructura

| Campo | Valor |
|---|---|
| Actor principal | Todos los roles |
| Prioridad | **Alta — foco principal de la primera etapa** |
| Depende de | CU-02, CU-05. CU-03, CU-10 y CU-08 suman secciones cuando existan |

## Descripción
Pantalla de inicio: en un vistazo, qué hay, qué está mal y qué documentación se movió.
Se construye por etapas: cada módulo nuevo agrega su sección sin romper las anteriores.

## Secciones

### Etapa 1 (con CU-02 y CU-05)
- **Tarjetas de inventario**: total de servidores, por entorno (prod/staging/dev) y por ubicación (on-premise/cloud).
- **Tarjetas de documentación**: total de documentos, por categoría, actualizados en los últimos 7 días.
- **Últimos documentos editados** (5): título, autor, fecha → click abre el documento.
- **Servidores sin documentación vinculada**: lista corta, sirve para detectar huecos.
- **Gráfico**: documentos por categoría (barras, Recharts).

### Etapa 2 (al sumar CU-03 y CU-10)
- Servicios UP / DEGRADED / DOWN (tarjetas de color).
- Incidentes abiertos por severidad.
- Gráficos de uptime % y latencia de las últimas 24 h.

### Etapa 3 (al sumar CU-08)
- Notificaciones no leídas del usuario.

## Reglas de negocio
- RN-1 Un único endpoint agregado por etapa para no hacer N requests.
- RN-2 Las secciones de módulos todavía no implementados no se muestran (no dan error).
- RN-3 Cada tarjeta es clickeable y navega al listado ya filtrado (ej. "12 servidores en prod" → `/inventario?environment=prod`).
- RN-4 Uptime % = chequeos UP o DEGRADED / total de chequeos en la ventana × 100.

## Endpoints
| Método | Ruta | Rol mínimo | Notas |
|---|---|---|---|
| GET | /api/dashboard/summary | soporte | contadores y listas cortas; campos opcionales por módulo |
| GET | /api/dashboard/timeseries | soporte | etapa 2: `window=24h`, buckets de 15 min |

## UI
- Grilla responsive: tarjetas arriba, listas y gráficos abajo.
- Refresco automático cada 30 s (`refetchInterval`) sin recargar la página.
- Colores de estado consistentes (ver frontend/AGENTS.md).
- Modo claro/oscuro.

## Criterios de aceptación
- [ ] CA-1 Los números coinciden con una consulta directa a la base.
- [ ] CA-2 Al crear un documento nuevo, aparece en "Últimos editados" sin recargar la página (en ≤ 30 s).
- [ ] CA-3 Click en una tarjeta navega al listado filtrado.
- [ ] CA-4 Sin CU-03 implementado, el dashboard carga sin errores.
