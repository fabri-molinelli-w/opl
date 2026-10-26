import { useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import {
  getListDocumentsQueryKey,
  useDeleteDocument,
  useGetDocument,
  useGetDocumentVersion,
  useListDocumentVersions,
} from '@/api/gen'
import { isApiError } from '@/api/http'
import { hasRole } from '@/auth/roles'
import { useAuth } from '@/auth/useAuth'
import { Markdown } from '@/components/Markdown'
import { CATEGORY_LABELS, formatDate } from './categories'

export const DocumentDetailPage = () => {
  const { id = '' } = useParams()
  const { user } = useAuth()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const canDelete = hasRole(user?.role, 'sysadmin')
  const [selectedVersion, setSelectedVersion] = useState<number | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const docResult = useGetDocument(id)
  const versionsResult = useListDocumentVersions(id)
  const versionResult = useGetDocumentVersion(id, selectedVersion ?? 0, {
    query: { enabled: selectedVersion !== null },
  })
  const deleteDocument = useDeleteDocument()

  const doc = docResult.data?.status === 200 ? docResult.data.data : null
  const versions = versionsResult.data?.status === 200 ? versionsResult.data.data : []
  const oldVersion =
    selectedVersion !== null && versionResult.data?.status === 200 ? versionResult.data.data : null

  if (docResult.isLoading) {
    return <p className="text-sm text-slate-500">Cargando documento…</p>
  }
  if (doc === null) {
    return (
      <section className="space-y-3">
        <p role="alert" className="text-sm text-red-700">
          El documento no existe o fue eliminado.
        </p>
        <Link to="/documentacion" className="text-sm text-sky-700 underline">
          Volver a la documentación
        </Link>
      </section>
    )
  }

  const remove = async () => {
    if (!window.confirm(`¿Eliminar el documento "${doc.title}"?`)) {
      return
    }
    setActionError(null)
    try {
      await deleteDocument.mutateAsync({ id: doc.id })
      await queryClient.invalidateQueries({ queryKey: getListDocumentsQueryKey() })
      void navigate('/documentacion')
    } catch (cause) {
      setActionError(isApiError(cause) ? cause.message : 'No se pudo eliminar el documento')
    }
  }

  return (
    <section className="grid gap-6 lg:grid-cols-[1fr_18rem]">
      <article className="space-y-4 rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <Link to="/documentacion" className="text-sm text-sky-700 hover:underline">
          ← Documentación
        </Link>
        <header className="space-y-2">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <h1 className="text-2xl font-semibold">{doc.title}</h1>
            <div className="flex gap-2">
              <Link
                to={`/documentacion/${doc.id}/editar`}
                className="rounded-md bg-slate-900 px-3 py-1.5 text-sm font-medium text-white hover:bg-slate-800"
              >
                Editar
              </Link>
              {canDelete && (
                <button
                  type="button"
                  onClick={() => void remove()}
                  className="rounded-md border border-red-300 px-3 py-1.5 text-sm font-medium text-red-700 hover:bg-red-50"
                >
                  Eliminar
                </button>
              )}
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2 text-xs text-slate-500">
            <span className="rounded-full bg-slate-100 px-2 py-0.5 font-medium text-slate-700">
              {CATEGORY_LABELS[doc.category]}
            </span>
            {doc.tags.map((t) => (
              <Link
                key={t}
                to={`/documentacion?tag=${encodeURIComponent(t)}`}
                className="rounded bg-sky-50 px-1.5 py-0.5 text-sky-800 hover:bg-sky-100"
              >
                #{t}
              </Link>
            ))}
            <span>
              v{doc.version} · creado por {doc.createdByName} · editado por {doc.updatedByName} el{' '}
              {formatDate(doc.updatedAt)}
            </span>
          </div>
        </header>
        {actionError !== null && (
          <p role="alert" className="text-sm text-red-700">
            {actionError}
          </p>
        )}

        {oldVersion !== null ? (
          <div className="space-y-3">
            <div className="flex items-center justify-between rounded-md bg-amber-50 px-3 py-2 text-sm text-amber-900">
              <span>
                Estás viendo la versión {oldVersion.version} ({oldVersion.editedByName},{' '}
                {formatDate(oldVersion.createdAt)}).
              </span>
              <button type="button" onClick={() => setSelectedVersion(null)} className="underline">
                Volver a la versión actual
              </button>
            </div>
            <h2 className="text-xl font-semibold">{oldVersion.title}</h2>
            <Markdown content={oldVersion.contentMd} />
          </div>
        ) : (
          <Markdown content={doc.contentMd} />
        )}
      </article>

      <aside className="space-y-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">Historial</h2>
        <ul className="space-y-1">
          {versions.map((v) => {
            const isCurrent = v.version === doc.version
            const isSelected = selectedVersion === v.version || (selectedVersion === null && isCurrent)
            return (
              <li key={v.version}>
                <button
                  type="button"
                  onClick={() => setSelectedVersion(isCurrent ? null : v.version)}
                  className={`w-full rounded-md px-2 py-1.5 text-left text-sm ${
                    isSelected ? 'bg-slate-100 font-medium' : 'hover:bg-slate-50'
                  }`}
                >
                  v{v.version}
                  {isCurrent ? ' (actual)' : ''} · {v.editedByName}
                  <span className="block text-xs text-slate-500">{formatDate(v.createdAt)}</span>
                </button>
              </li>
            )
          })}
        </ul>
      </aside>
    </section>
  )
}
