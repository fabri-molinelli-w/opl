import { Link, useSearchParams } from 'react-router-dom'
import { useListDocumentTags, useListDocuments } from '@/api/gen'
import type { DocumentCategory, DocumentSort, SortOrder } from '@/api/gen/model'
import { CATEGORIES, CATEGORY_LABELS, formatDate } from './categories'

const PAGE_SIZE = 20

const SORT_OPTIONS: { value: string; label: string; sort: DocumentSort; order: SortOrder }[] = [
  { value: 'updatedAt-desc', label: 'Última edición', sort: 'updatedAt', order: 'desc' },
  { value: 'updatedAt-asc', label: 'Edición más antigua', sort: 'updatedAt', order: 'asc' },
  { value: 'title-asc', label: 'Título (A-Z)', sort: 'title', order: 'asc' },
  { value: 'title-desc', label: 'Título (Z-A)', sort: 'title', order: 'desc' },
]

const positiveInt = (raw: string | null, fallback: number): number => {
  const value = Number(raw)
  return Number.isInteger(value) && value > 0 ? value : fallback
}

export const DocumentsPage = () => {
  const [searchParams, setSearchParams] = useSearchParams()
  const page = positiveInt(searchParams.get('page'), 1)
  const q = searchParams.get('q') ?? ''
  const category = searchParams.get('category') ?? ''
  const tag = searchParams.get('tag') ?? ''
  const sortValue = searchParams.get('sort') ?? 'updatedAt-desc'
  const sortOption = SORT_OPTIONS.find((o) => o.value === sortValue) ?? SORT_OPTIONS[0]

  const listResult = useListDocuments({
    page,
    pageSize: PAGE_SIZE,
    q: q === '' ? undefined : q,
    category: category === '' ? undefined : (category as DocumentCategory),
    tag: tag === '' ? undefined : tag,
    sort: sortOption.sort,
    order: sortOption.order,
  })
  const tagsResult = useListDocumentTags()
  const allTags = tagsResult.data?.status === 200 ? tagsResult.data.data : []

  const response = listResult.data
  const data = response !== undefined && response.status === 200 ? response.data : null
  const documents = data?.items ?? []
  const totalPages = data === null ? 1 : Math.max(1, Math.ceil(data.total / data.pageSize))

  const updateFilter = (name: string, value: string) => {
    const next = new URLSearchParams(searchParams)
    if (value === '') {
      next.delete(name)
    } else {
      next.set(name, value)
    }
    next.set('page', '1')
    setSearchParams(next)
  }

  const goToPage = (nextPage: number) => {
    const next = new URLSearchParams(searchParams)
    next.set('page', String(nextPage))
    setSearchParams(next)
  }

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Documentación</h1>
        <Link
          to="/documentacion/nuevo"
          className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800"
        >
          Nuevo documento
        </Link>
      </div>

      <div className="space-y-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
        <input
          aria-label="Buscar documentos"
          placeholder="Buscar en títulos, contenido y etiquetas (ej. vpn, nginx, backup)"
          value={q}
          onChange={(event) => updateFilter('q', event.target.value)}
          className="w-full rounded-md border border-slate-300 px-4 py-3 text-base"
        />
        <div className="grid gap-3 md:grid-cols-3">
          <select
            aria-label="Categoría"
            value={category}
            onChange={(event) => updateFilter('category', event.target.value)}
            className="rounded-md border border-slate-300 px-3 py-2 text-sm"
          >
            <option value="">Todas las categorías</option>
            {CATEGORIES.map((c) => (
              <option key={c} value={c}>
                {CATEGORY_LABELS[c]}
              </option>
            ))}
          </select>
          <select
            aria-label="Etiqueta"
            value={tag}
            onChange={(event) => updateFilter('tag', event.target.value)}
            className="rounded-md border border-slate-300 px-3 py-2 text-sm"
          >
            <option value="">Todas las etiquetas</option>
            {allTags.map((t) => (
              <option key={t} value={t}>
                {t}
              </option>
            ))}
          </select>
          <select
            aria-label="Ordenar por"
            value={sortOption.value}
            onChange={(event) => updateFilter('sort', event.target.value)}
            className="rounded-md border border-slate-300 px-3 py-2 text-sm"
          >
            {SORT_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
        </div>
      </div>

      {listResult.isLoading && <p className="text-sm text-slate-500">Cargando documentos…</p>}
      {listResult.isError && (
        <p role="alert" className="text-sm text-red-700">
          No se pudieron cargar los documentos.
        </p>
      )}

      {data !== null && documents.length === 0 && (
        <p className="rounded-xl border border-dashed border-slate-300 bg-white p-6 text-center text-sm text-slate-500">
          No hay documentos que coincidan con la búsqueda.
        </p>
      )}

      <ul className="grid gap-3">
        {documents.map((doc) => (
          <li key={doc.id} className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
            <div className="flex flex-wrap items-baseline justify-between gap-2">
              <Link to={`/documentacion/${doc.id}`} className="text-lg font-medium text-slate-900 hover:underline">
                {doc.title}
              </Link>
              <span className="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-medium text-slate-700">
                {CATEGORY_LABELS[doc.category]}
              </span>
            </div>
            <div className="mt-2 flex flex-wrap items-center gap-2 text-xs text-slate-500">
              {doc.tags.map((t) => (
                <button
                  key={t}
                  type="button"
                  onClick={() => updateFilter('tag', t)}
                  className="rounded bg-sky-50 px-1.5 py-0.5 text-sky-800 hover:bg-sky-100"
                >
                  #{t}
                </button>
              ))}
              <span>
                v{doc.version} · editado por {doc.updatedByName} el {formatDate(doc.updatedAt)}
              </span>
            </div>
          </li>
        ))}
      </ul>

      {data !== null && totalPages > 1 && (
        <div className="flex items-center justify-center gap-3 text-sm">
          <button
            type="button"
            disabled={page <= 1}
            onClick={() => goToPage(page - 1)}
            className="rounded-md border border-slate-300 px-3 py-1 disabled:opacity-40"
          >
            Anterior
          </button>
          <span>
            Página {page} de {totalPages}
          </span>
          <button
            type="button"
            disabled={page >= totalPages}
            onClick={() => goToPage(page + 1)}
            className="rounded-md border border-slate-300 px-3 py-1 disabled:opacity-40"
          >
            Siguiente
          </button>
        </div>
      )}
    </section>
  )
}
