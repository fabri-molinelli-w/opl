import { useState, type FormEvent } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { getListDeploysQueryKey, useCreateDeploy, useListDeploys, useListServices } from '@/api/gen'
import type { Deploy, DeployResult, Environment } from '@/api/gen/model'
import { hasRole } from '@/auth/roles'
import { useAuth } from '@/auth/useAuth'
import { formErrorMessage } from '@/features/inventory/formErrors'
import { formatDateTime } from '@/features/incidents/labels'

const RESULT_STYLES: Record<DeployResult, string> = {
  success: 'bg-emerald-100 text-emerald-800',
  failed: 'bg-red-100 text-red-800',
  rollback: 'bg-amber-100 text-amber-800',
}

export const DeployTable = ({ items, empty }: { items: Deploy[]; empty: string }) => (
  <table className="w-full text-left text-sm">
    <thead className="text-slate-500">
      <tr>
        <th className="py-2">Fecha</th>
        <th>Servicio</th>
        <th>Versión</th>
        <th>Ambiente</th>
        <th>Resultado</th>
        <th>Autor</th>
        <th>Origen</th>
      </tr>
    </thead>
    <tbody>
      {items.length === 0 && (
        <tr>
          <td colSpan={7} className="py-6 text-center text-slate-500">
            {empty}
          </td>
        </tr>
      )}
      {items.map((d) => (
        <tr key={d.id} className="border-t border-slate-100" title={d.notes}>
          <td className="py-2">{formatDateTime(d.deployedAt)}</td>
          <td className="font-medium">{d.serviceName}</td>
          <td>{d.version}</td>
          <td>{d.environment}</td>
          <td>
            <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${RESULT_STYLES[d.result]}`}>
              {d.result}
            </span>
          </td>
          <td>{d.author || '—'}</td>
          <td>{d.source}</td>
        </tr>
      ))}
    </tbody>
  </table>
)

/** CU-07: historial de deploys con filtros y alta manual (sysadmin). */
export const DeploysPage = () => {
  const { user } = useAuth()
  const canWrite = hasRole(user?.role, 'sysadmin')
  const queryClient = useQueryClient()
  const [serviceId, setServiceId] = useState('')
  const [environment, setEnvironment] = useState('')
  const services = useListServices({ pageSize: 100 })
  const serviceItems = services.data?.status === 200 ? services.data.data.items : []
  const list = useListDeploys({
    serviceId: serviceId === '' ? undefined : serviceId,
    environment: environment === '' ? undefined : (environment as Environment),
  })
  const items = list.data?.status === 200 && Array.isArray(list.data.data) ? list.data.data : []

  const create = useCreateDeploy()
  const [form, setForm] = useState({
    serviceId: '',
    version: '',
    result: 'success' as DeployResult,
    notes: '',
  })
  const [error, setError] = useState('')

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    setError('')
    try {
      await create.mutateAsync({ data: { ...form, author: user?.name ?? '' } })
      setForm({ ...form, version: '', notes: '' })
      await queryClient.invalidateQueries({
        queryKey: getListDeploysQueryKey(),
      })
    } catch (cause) {
      setError(formErrorMessage(cause, 'No se pudo registrar el deploy'))
    }
  }

  return (
    <section className="space-y-6">
      <h1 className="text-2xl font-semibold">Deploys</h1>

      {canWrite && (
        <form
          onSubmit={(e) => void submit(e)}
          className="flex flex-wrap items-end gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm"
        >
          <select
            required
            aria-label="Servicio del deploy"
            value={form.serviceId}
            onChange={(e) => setForm({ ...form, serviceId: e.target.value })}
            className="rounded-md border border-slate-300 px-3 py-2 text-sm"
          >
            <option value="">Servicio…</option>
            {serviceItems.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
          <input
            required
            aria-label="Versión"
            placeholder="v1.4.2"
            value={form.version}
            onChange={(e) => setForm({ ...form, version: e.target.value })}
            className="rounded-md border border-slate-300 px-3 py-2 text-sm"
          />
          <select
            aria-label="Resultado"
            value={form.result}
            onChange={(e) => setForm({ ...form, result: e.target.value as DeployResult })}
            className="rounded-md border border-slate-300 px-3 py-2 text-sm"
          >
            <option value="success">success</option>
            <option value="failed">failed</option>
            <option value="rollback">rollback</option>
          </select>
          <input
            aria-label="Notas"
            placeholder="Notas (opcional)"
            value={form.notes}
            onChange={(e) => setForm({ ...form, notes: e.target.value })}
            className="flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm"
          />
          <button
            type="submit"
            disabled={create.isPending}
            className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:opacity-50"
          >
            Registrar deploy
          </button>
          {error !== '' && <p className="w-full text-sm text-red-600">{error}</p>}
        </form>
      )}

      <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
        <div className="mb-4 flex gap-3">
          <select
            aria-label="Filtrar por servicio"
            value={serviceId}
            onChange={(e) => setServiceId(e.target.value)}
            className="rounded-md border border-slate-300 px-3 py-2 text-sm"
          >
            <option value="">Todos los servicios</option>
            {serviceItems.map((s) => (
              <option key={s.id} value={s.id}>
                {s.name}
              </option>
            ))}
          </select>
          <select
            aria-label="Filtrar por ambiente"
            value={environment}
            onChange={(e) => setEnvironment(e.target.value)}
            className="rounded-md border border-slate-300 px-3 py-2 text-sm"
          >
            <option value="">Todos los ambientes</option>
            <option value="prod">prod</option>
            <option value="staging">staging</option>
            <option value="dev">dev</option>
          </select>
        </div>
        <DeployTable items={items} empty={list.isLoading ? 'Cargando…' : 'No hay deploys registrados.'} />
      </div>
    </section>
  )
}

/** CU-07 RN-4: deploys del servicio en las 24 h previas a la apertura del incidente. */
export const RecentDeploys = ({ serviceId, openedAt }: { serviceId: string; openedAt: string }) => {
  const to = new Date(openedAt)
  const from = new Date(to.getTime() - 24 * 60 * 60 * 1000)
  const list = useListDeploys({
    serviceId,
    from: from.toISOString(),
    to: to.toISOString(),
  })
  const items = list.data?.status === 200 && Array.isArray(list.data.data) ? list.data.data : []
  return (
    <div className="space-y-3 rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 className="font-semibold">Deploys en las 24 h previas</h2>
      <DeployTable items={items} empty="Sin deploys del servicio en las 24 h anteriores al incidente." />
    </div>
  )
}
