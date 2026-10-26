import { useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import {
  getGetServerQueryKey,
  getListServersQueryKey,
  useDeleteServer,
  useDeleteService,
  useGetServer,
} from '@/api/gen'
import type { Service, ServiceStatus } from '@/api/gen/model'
import { isApiError } from '@/api/http'
import { hasRole } from '@/auth/roles'
import { useAuth } from '@/auth/useAuth'
import { ServerForm } from './ServerForm'
import { ServiceChecks } from './ServiceChecks'
import { ServiceForm } from './ServiceForm'

const STATUS_STYLES: Record<ServiceStatus, string> = {
  UP: 'bg-green-100 text-green-700',
  DEGRADED: 'bg-amber-100 text-amber-700',
  DOWN: 'bg-red-100 text-red-700',
  UNKNOWN: 'bg-slate-200 text-slate-600',
}

const formatDate = (value: string): string =>
  new Intl.DateTimeFormat('es-AR', {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value))

export const ServerDetailPage = () => {
  const { user } = useAuth()
  const canWrite = hasRole(user?.role, 'sysadmin')
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const getServer = useGetServer(id ?? '', {
    query: { enabled: id !== undefined, refetchInterval: 10_000 },
  })
  const [checksFor, setChecksFor] = useState<Service | null>(null)
  const deleteServer = useDeleteServer()
  const deleteService = useDeleteService()
  const [editingServer, setEditingServer] = useState(false)
  const [serviceForm, setServiceForm] = useState<
    { mode: 'create' } | { mode: 'edit'; service: Service } | null
  >(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const response = getServer.data
  const server =
    response !== undefined && response.status === 200 ? response.data : null

  const refresh = async () => {
    if (id !== undefined) {
      await queryClient.invalidateQueries({ queryKey: getGetServerQueryKey(id) })
    }
    await queryClient.invalidateQueries({ queryKey: getListServersQueryKey() })
  }

  const removeServer = async () => {
    if (server === null || !window.confirm(`¿Eliminar ${server.hostname}?`)) {
      return
    }
    setActionError(null)
    try {
      await deleteServer.mutateAsync({ id: server.id })
      await refresh()
      navigate('/inventario')
    } catch (cause) {
      setActionError(
        isApiError(cause)
          ? cause.message
          : 'No se pudo eliminar el servidor',
      )
    }
  }

  const removeService = async (service: Service) => {
    if (!window.confirm(`¿Eliminar el servicio ${service.name}?`)) {
      return
    }
    setActionError(null)
    try {
      await deleteService.mutateAsync({ id: service.id })
      await refresh()
    } catch (cause) {
      setActionError(
        isApiError(cause)
          ? cause.message
          : 'No se pudo eliminar el servicio',
      )
    }
  }

  if (id === undefined) {
    return <p className="text-sm text-red-600">Servidor inexistente.</p>
  }

  return (
    <section className="space-y-6">
      <Link to="/inventario" className="text-sm font-medium text-slate-600 hover:underline">
        ← Volver al inventario
      </Link>

      {getServer.isPending && (
        <p className="text-sm text-slate-500">Cargando servidor…</p>
      )}

      {getServer.isError && (
        <p role="alert" className="text-sm text-red-600">
          {isApiError(getServer.error)
            ? getServer.error.message
            : 'No se pudo cargar el servidor'}
        </p>
      )}

      {actionError !== null && (
        <p role="alert" className="text-sm text-red-600">
          {actionError}
        </p>
      )}

      {server !== null && (
        <>
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <h1 className="text-2xl font-semibold">{server.hostname}</h1>
              <p className="text-sm text-slate-500">{server.ip}</p>
            </div>
            {canWrite && (
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setEditingServer(true)}
                  className="rounded-md border border-slate-300 px-3 py-2 text-sm font-medium"
                >
                  Editar servidor
                </button>
                <button
                  type="button"
                  onClick={() => void removeServer()}
                  className="rounded-md bg-red-600 px-3 py-2 text-sm font-medium text-white hover:bg-red-500"
                >
                  Eliminar servidor
                </button>
              </div>
            )}
          </div>

          {editingServer && (
            <ServerForm
              server={server}
              onClose={() => setEditingServer(false)}
            />
          )}

          <dl className="grid gap-4 rounded-xl border border-slate-200 bg-white p-4 text-sm shadow-sm md:grid-cols-3">
            <Detail label="Environment" value={server.environment} />
            <Detail label="Location" value={server.location} />
            <Detail label="Sistema operativo" value={server.os || '—'} />
            <Detail label="Responsable" value={server.owner || '—'} />
            <Detail label="Tags" value={server.tags.join(', ') || '—'} />
            <Detail label="Actualizado" value={formatDate(server.updatedAt)} />
          </dl>

          <div className="flex items-center justify-between">
            <h2 className="text-lg font-semibold">Servicios</h2>
            {canWrite && (
              <button
                type="button"
                onClick={() => setServiceForm({ mode: 'create' })}
                className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800"
              >
                Nuevo servicio
              </button>
            )}
          </div>

          {serviceForm !== null && (
            <ServiceForm
              serverId={server.id}
              service={
                serviceForm.mode === 'edit' ? serviceForm.service : undefined
              }
              onClose={() => setServiceForm(null)}
            />
          )}

          <div className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
            <table className="w-full text-left text-sm">
              <thead className="bg-slate-50 text-slate-500">
                <tr>
                  <th className="px-4 py-3 font-medium">Nombre</th>
                  <th className="px-4 py-3 font-medium">Tipo</th>
                  <th className="px-4 py-3 font-medium">Puerto</th>
                  <th className="px-4 py-3 font-medium">Check</th>
                  <th className="px-4 py-3 font-medium">Estado</th>
                  <th className="px-4 py-3 font-medium">Monitoreo</th>
                  {canWrite && (
                    <th className="px-4 py-3 font-medium">Acciones</th>
                  )}
                </tr>
              </thead>
              <tbody>
                {server.services?.map((service) => (
                  <tr key={service.id} className="border-t border-slate-100">
                    <td className="px-4 py-3 font-medium">{service.name}</td>
                    <td className="px-4 py-3">{service.type}</td>
                    <td className="px-4 py-3">{service.port}</td>
                    <td className="px-4 py-3">
                      <span className="text-slate-500">{service.checkType}</span>{' '}
                      {service.checkTarget}
                    </td>
                    <td className="px-4 py-3">
                      <span
                        className={`rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_STYLES[service.currentStatus]}`}
                      >
                        {service.currentStatus}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <button
                        type="button"
                        onClick={() => setChecksFor(checksFor?.id === service.id ? null : service)}
                        className="text-sm font-medium text-sky-700 hover:underline"
                      >
                        {checksFor?.id === service.id ? 'Ocultar' : 'Historial'}
                      </button>
                    </td>
                    {canWrite && (
                      <td className="px-4 py-3">
                        <div className="flex gap-3">
                          <button
                            type="button"
                            onClick={() =>
                              setServiceForm({ mode: 'edit', service })
                            }
                            className="text-sm font-medium text-slate-700 hover:underline"
                          >
                            Editar
                          </button>
                          <button
                            type="button"
                            onClick={() => void removeService(service)}
                            className="text-sm font-medium text-red-600 hover:underline"
                          >
                            Eliminar
                          </button>
                        </div>
                      </td>
                    )}
                  </tr>
                ))}
              </tbody>
            </table>
            {(server.services ?? []).length === 0 && (
              <p className="px-4 py-6 text-sm text-slate-500">
                Este servidor todavía no tiene servicios.
              </p>
            )}
          </div>
          {checksFor !== null && (
            <ServiceChecks
              key={checksFor.id}
              serviceId={checksFor.id}
              serviceName={checksFor.name}
              serverId={server.id}
              canRun={canWrite}
            />
          )}
        </>
      )}
    </section>
  )
}

const Detail = ({ label, value }: { label: string; value: string }) => (
  <div>
    <dt className="text-xs uppercase tracking-wide text-slate-500">{label}</dt>
    <dd className="mt-1 text-slate-800">{value}</dd>
  </div>
)
