import { useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import {
  getListServersQueryKey,
  useDeleteServer,
  useListServers,
} from '@/api/gen'
import type { Environment, Server } from '@/api/gen/model'
import { isApiError } from '@/api/http'
import { hasRole } from '@/auth/roles'
import { useAuth } from '@/auth/useAuth'
import { ServerForm } from './ServerForm'

const ENVIRONMENTS: Environment[] = ['prod', 'staging', 'dev']
const PAGE_SIZE = 20

export const InventoryPage = () => {
  const { user } = useAuth()
  const canWrite = hasRole(user?.role, 'sysadmin')
  const [searchParams, setSearchParams] = useSearchParams()
  const queryClient = useQueryClient()
  const deleteServer = useDeleteServer()
  const [creating, setCreating] = useState(false)
  const [editing, setEditing] = useState<Server | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const page = positiveInt(searchParams.get('page'), 1)
  const q = searchParams.get('q') ?? ''
  const environment = searchParams.get('environment') ?? ''
  const os = searchParams.get('os') ?? ''
  const tag = searchParams.get('tag') ?? ''

  const listResult = useListServers({
    page,
    pageSize: PAGE_SIZE,
    q: q === '' ? undefined : q,
    environment: environment === '' ? undefined : (environment as Environment),
    os: os === '' ? undefined : os,
    tag: tag === '' ? undefined : tag,
  })
  const response = listResult.data
  const data =
    response !== undefined && response.status === 200 ? response.data : null
  const servers = data?.items ?? []
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

  const removeServer = async (server: Server) => {
    if (!window.confirm(`¿Eliminar el servidor ${server.hostname}?`)) {
      return
    }
    setActionError(null)
    try {
      await deleteServer.mutateAsync({ id: server.id })
      await queryClient.invalidateQueries({ queryKey: getListServersQueryKey() })
    } catch (cause) {
      setActionError(
        isApiError(cause)
          ? cause.message
          : 'No se pudo eliminar el servidor',
      )
    }
  }

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Inventario</h1>
        {canWrite && (
          <button
            type="button"
            onClick={() => {
              setEditing(null)
              setCreating(true)
            }}
            className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800"
          >
            Nuevo servidor
          </button>
        )}
      </div>

      {(creating || editing !== null) && (
        <ServerForm
          server={editing ?? undefined}
          onClose={() => {
            setCreating(false)
            setEditing(null)
          }}
        />
      )}

      <div className="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-4">
        <input
          aria-label="Buscar"
          placeholder="Buscar por hostname, IP o responsable"
          value={q}
          onChange={(event) => updateFilter('q', event.target.value)}
          className="rounded-md border border-slate-300 px-3 py-2 text-sm"
        />
        <select
          aria-label="Environment"
          value={environment}
          onChange={(event) => updateFilter('environment', event.target.value)}
          className="rounded-md border border-slate-300 px-3 py-2 text-sm"
        >
          <option value="">Todos los environments</option>
          {ENVIRONMENTS.map((option) => (
            <option key={option} value={option}>
              {option}
            </option>
          ))}
        </select>
        <input
          aria-label="Sistema operativo"
          placeholder="Sistema operativo"
          value={os}
          onChange={(event) => updateFilter('os', event.target.value)}
          className="rounded-md border border-slate-300 px-3 py-2 text-sm"
        />
        <input
          aria-label="Tag"
          placeholder="Tag"
          value={tag}
          onChange={(event) => updateFilter('tag', event.target.value)}
          className="rounded-md border border-slate-300 px-3 py-2 text-sm"
        />
      </div>

      {actionError !== null && (
        <p role="alert" className="text-sm text-red-600">
          {actionError}
        </p>
      )}

      {listResult.isPending && (
        <p className="text-sm text-slate-500">Cargando servidores…</p>
      )}

      {listResult.isError && (
        <p role="alert" className="text-sm text-red-600">
          {isApiError(listResult.error)
            ? listResult.error.message
            : 'No se pudieron cargar los servidores'}
        </p>
      )}

      {!listResult.isPending && !listResult.isError && (
        <div className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-50 text-slate-500">
              <tr>
                <th className="px-4 py-3 font-medium">Hostname</th>
                <th className="px-4 py-3 font-medium">IP</th>
                <th className="px-4 py-3 font-medium">Environment</th>
                <th className="px-4 py-3 font-medium">SO</th>
                <th className="px-4 py-3 font-medium">Tags</th>
                <th className="px-4 py-3 font-medium">Acciones</th>
              </tr>
            </thead>
            <tbody>
              {servers.map((server) => (
                <tr key={server.id} className="border-t border-slate-100">
                  <td className="px-4 py-3 font-medium">{server.hostname}</td>
                  <td className="px-4 py-3">{server.ip}</td>
                  <td className="px-4 py-3">{server.environment}</td>
                  <td className="px-4 py-3">{server.os === '' ? '—' : server.os}</td>
                  <td className="px-4 py-3">
                    <div className="flex flex-wrap gap-1">
                      {server.tags.map((item) => (
                        <span
                          key={item}
                          className="rounded-full bg-slate-100 px-2 py-0.5 text-xs text-slate-600"
                        >
                          {item}
                        </span>
                      ))}
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex gap-3">
                      <Link
                        to={`/inventario/${server.id}`}
                        className="text-sm font-medium text-slate-700 hover:underline"
                      >
                        Ver ficha
                      </Link>
                      {canWrite && (
                        <>
                          <button
                            type="button"
                            onClick={() => {
                              setCreating(false)
                              setEditing(server)
                            }}
                            className="text-sm font-medium text-slate-700 hover:underline"
                          >
                            Editar
                          </button>
                          <button
                            type="button"
                            onClick={() => void removeServer(server)}
                            className="text-sm font-medium text-red-600 hover:underline"
                          >
                            Eliminar
                          </button>
                        </>
                      )}
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {servers.length === 0 && (
            <p className="px-4 py-6 text-sm text-slate-500">
              No hay servidores para los filtros seleccionados.
            </p>
          )}
          <div className="flex items-center justify-between border-t border-slate-100 px-4 py-3 text-sm">
            <span className="text-slate-500">
              {data?.total ?? 0} servidores · página {page} de {totalPages}
            </span>
            <div className="flex gap-2">
              <button
                type="button"
                disabled={page <= 1}
                onClick={() => goToPage(page - 1)}
                className="rounded-md border border-slate-300 px-3 py-1 disabled:opacity-40"
              >
                Anterior
              </button>
              <button
                type="button"
                disabled={page >= totalPages}
                onClick={() => goToPage(page + 1)}
                className="rounded-md border border-slate-300 px-3 py-1 disabled:opacity-40"
              >
                Siguiente
              </button>
            </div>
          </div>
        </div>
      )}
    </section>
  )
}

const positiveInt = (raw: string | null, fallback: number): number => {
  if (raw === null) return fallback
  const value = Number(raw)
  return Number.isInteger(value) && value > 0 ? value : fallback
}
