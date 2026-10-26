import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import {
  getGetDocumentQueryKey,
  getListDocumentTagsQueryKey,
  getListDocumentVersionsQueryKey,
  getListDocumentsQueryKey,
  useCreateDocument,
  useGetDocument,
  useListDocumentTags,
  useUpdateDocument,
} from '@/api/gen'
import type { Document, DocumentCategory } from '@/api/gen/model'
import { Markdown } from '@/components/Markdown'
import { fieldErrors, formErrorMessage } from '@/features/inventory/formErrors'
import { CATEGORIES, CATEGORY_LABELS } from './categories'

const TEMPLATE = `## Objetivo

## Requisitos

## Pasos

1.

\`\`\`bash
# comando
\`\`\`

## Verificación
`

const parseTags = (raw: string): string[] =>
  raw
    .split(',')
    .map((tag) => tag.trim())
    .filter((tag) => tag.length > 0)

/** Ruta /documentacion/:id/editar: carga el documento y monta el formulario. */
export const DocumentEditorPage = () => {
  const { id } = useParams()
  const docResult = useGetDocument(id ?? '', { query: { enabled: id !== undefined } })

  if (id === undefined) {
    return <DocumentForm />
  }
  if (docResult.isLoading) {
    return <p className="text-sm text-slate-500">Cargando documento…</p>
  }
  const doc = docResult.data?.status === 200 ? docResult.data.data : null
  if (doc === null) {
    return (
      <p role="alert" className="text-sm text-red-700">
        El documento no existe o fue eliminado.
      </p>
    )
  }
  return <DocumentForm key={`${doc.id}-${doc.version}`} document={doc} />
}

const DocumentForm = ({ document }: { document?: Document }) => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const createDocument = useCreateDocument()
  const updateDocument = useUpdateDocument()
  const tagsResult = useListDocumentTags()
  const knownTags = tagsResult.data?.status === 200 ? tagsResult.data.data : []

  const [title, setTitle] = useState(document?.title ?? '')
  const [category, setCategory] = useState<DocumentCategory>(document?.category ?? 'procedimiento')
  const [tags, setTags] = useState(document?.tags.join(', ') ?? '')
  const [content, setContent] = useState(document?.contentMd ?? TEMPLATE)
  const [error, setError] = useState<string | null>(null)
  const [errors, setErrors] = useState<Record<string, string>>({})
  const [saving, setSaving] = useState(false)

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError(null)
    setErrors({})
    setSaving(true)
    const data = { title, category, tags: parseTags(tags), contentMd: content }
    try {
      let savedId: string
      if (document === undefined) {
        const result = await createDocument.mutateAsync({ data })
        savedId = result.status === 201 ? result.data.id : ''
      } else {
        await updateDocument.mutateAsync({
          id: document.id,
          data: { ...data, version: document.version },
        })
        savedId = document.id
        await queryClient.invalidateQueries({ queryKey: getGetDocumentQueryKey(document.id) })
        await queryClient.invalidateQueries({ queryKey: getListDocumentVersionsQueryKey(document.id) })
      }
      await queryClient.invalidateQueries({ queryKey: getListDocumentsQueryKey() })
      await queryClient.invalidateQueries({ queryKey: getListDocumentTagsQueryKey() })
      void navigate(savedId === '' ? '/documentacion' : `/documentacion/${savedId}`)
    } catch (cause) {
      // El texto del editor se conserva siempre: nunca se pierde lo escrito.
      setErrors(fieldErrors(cause))
      setError(formErrorMessage(cause, 'No se pudo guardar el documento'))
    } finally {
      setSaving(false)
    }
  }

  const cancelTo = document === undefined ? '/documentacion' : `/documentacion/${document.id}`

  return (
    <form onSubmit={(event) => void submit(event)} className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">
          {document === undefined ? 'Nuevo documento' : `Editar: ${document.title}`}
        </h1>
        <div className="flex gap-2">
          <Link to={cancelTo} className="rounded-md border border-slate-300 px-3 py-2 text-sm hover:bg-slate-100">
            Cancelar
          </Link>
          <button
            type="submit"
            disabled={saving}
            className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:opacity-50"
          >
            {saving ? 'Guardando…' : 'Guardar'}
          </button>
        </div>
      </div>

      {error !== null && (
        <div role="alert" className="rounded-md border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-800">
          {error}
          {errors.version !== undefined && document !== undefined && (
            <span>
              {' '}
              <a href={`/documentacion/${document.id}`} target="_blank" rel="noopener noreferrer" className="underline">
                Ver la versión actual en otra pestaña
              </a>
              . Tus cambios siguen acá: copialos antes de recargar.
            </span>
          )}
        </div>
      )}

      <div className="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-3">
        <label className="space-y-1 text-sm md:col-span-3">
          <span className="font-medium">Título</span>
          <input
            value={title}
            onChange={(event) => setTitle(event.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
            placeholder="Ej. Configurar VPN en notebook"
          />
          {errors.title !== undefined && <span className="text-xs text-red-700">{errors.title}</span>}
        </label>
        <label className="space-y-1 text-sm">
          <span className="font-medium">Categoría</span>
          <select
            value={category}
            onChange={(event) => setCategory(event.target.value as DocumentCategory)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          >
            {CATEGORIES.map((c) => (
              <option key={c} value={c}>
                {CATEGORY_LABELS[c]}
              </option>
            ))}
          </select>
        </label>
        <label className="space-y-1 text-sm md:col-span-2">
          <span className="font-medium">Etiquetas (separadas por coma)</span>
          <input
            value={tags}
            onChange={(event) => setTags(event.target.value)}
            list="document-tags"
            className="w-full rounded-md border border-slate-300 px-3 py-2"
            placeholder="vpn, notebooks"
          />
          <datalist id="document-tags">
            {knownTags.map((t) => (
              <option key={t} value={t} />
            ))}
          </datalist>
          {errors.tags !== undefined && <span className="text-xs text-red-700">{errors.tags}</span>}
        </label>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <label className="space-y-1 text-sm">
          <span className="font-medium">Contenido (Markdown)</span>
          <textarea
            aria-label="Contenido"
            value={content}
            onChange={(event) => setContent(event.target.value)}
            className="h-[60vh] w-full rounded-md border border-slate-300 p-3 font-mono text-sm"
          />
          {errors.contentMd !== undefined && <span className="text-xs text-red-700">{errors.contentMd}</span>}
        </label>
        <div className="space-y-1 text-sm">
          <span className="font-medium">Vista previa</span>
          <div className="h-[60vh] overflow-y-auto rounded-md border border-slate-200 bg-white p-4">
            <Markdown content={content} />
          </div>
        </div>
      </div>
    </form>
  )
}
