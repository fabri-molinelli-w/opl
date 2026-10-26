import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { getGetServerQueryKey, getListServiceChecksQueryKey, useListServiceChecks, useRunServiceCheck } from '@/api/gen'
import type { HealthCheckStatus } from '@/api/gen/model'
import { isApiError } from '@/api/http'

const CHECK_STYLES: Record<HealthCheckStatus, string> = {
  UP: 'bg-green-100 text-green-700',
  DEGRADED: 'bg-amber-100 text-amber-700',
  DOWN: 'bg-red-100 text-red-700',
}

const formatTime = (value: string): string =>
  new Intl.DateTimeFormat('es-AR', { dateStyle: 'short', timeStyle: 'medium' }).format(new Date(value))

type Props = { serviceId: string; serviceName: string; serverId: string; canRun: boolean }

/** Historial de chequeos de un servicio + "Chequear ahora" (CU-03 RN-8, CA-3). */
export const ServiceChecks = ({ serviceId, serviceName, serverId, canRun }: Props) => {
  const queryClient = useQueryClient()
  const checks = useListServiceChecks(serviceId, { limit: 20 }, { query: { refetchInterval: 10_000 } })
  const runCheck = useRunServiceCheck()
  const [error, setError] = useState<string | null>(null)
  const items = checks.data?.status === 200 ? checks.data.data : []

  const run = async () => {
    setError(null)
    try {
      await runCheck.mutateAsync({ id: serviceId })
      await queryClient.invalidateQueries({ queryKey: getListServiceChecksQueryKey(serviceId) })
      await queryClient.invalidateQueries({ queryKey: getGetServerQueryKey(serverId) })
    } catch (cause) {
      setError(isApiError(cause) ? cause.message : 'No se pudo ejecutar el chequeo')
    }
  }

  return (
    <div className="space-y-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex items-center justify-between">
        <h3 className="font-semibold">Chequeos de {serviceName}</h3>
        {canRun && (
          <button
            type="button"
            onClick={() => void run()}
            disabled={runCheck.isPending}
            className="rounded-md bg-slate-900 px-3 py-1.5 text-sm font-medium text-white hover:bg-slate-800 disabled:opacity-50"
          >
            {runCheck.isPending ? 'Chequeando…' : 'Chequear ahora'}
          </button>
        )}
      </div>
      {error !== null && (
        <p role="alert" className="text-sm text-red-700">
          {error}
        </p>
      )}
      {items.length === 0 ? (
        <p className="text-sm text-slate-500">Todavía no hay chequeos registrados.</p>
      ) : (
        <table className="w-full text-left text-sm">
          <thead className="text-slate-500">
            <tr>
              <th className="py-1 font-medium">Fecha</th>
              <th className="py-1 font-medium">Estado</th>
              <th className="py-1 font-medium">Latencia</th>
              <th className="py-1 font-medium">HTTP</th>
              <th className="py-1 font-medium">Error</th>
            </tr>
          </thead>
          <tbody>
            {items.map((c) => (
              <tr key={c.id} className="border-t border-slate-100">
                <td className="py-1">{formatTime(c.checkedAt)}</td>
                <td className="py-1">
                  <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${CHECK_STYLES[c.status]}`}>
                    {c.status}
                  </span>
                </td>
                <td className="py-1">{c.responseMs !== undefined ? `${c.responseMs} ms` : '—'}</td>
                <td className="py-1">{c.httpStatus ?? '—'}</td>
                <td className="max-w-xs truncate py-1 text-slate-500" title={c.error}>
                  {c.error ?? ''}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
