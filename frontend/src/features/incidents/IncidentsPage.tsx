import { useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { getListIncidentsQueryKey, useCreateIncident, useListIncidents, useListServices } from '@/api/gen'
import type { IncidentStatus, Severity } from '@/api/gen/model'
import { fieldErrors, formErrorMessage } from '@/features/inventory/formErrors'
import { SEVERITIES, SEVERITY_STYLES, STATUSES, STATUS_LABELS, STATUS_STYLES, formatDateTime } from './labels'

const PAGE_SIZE = 20

export const IncidentsPage = () => {
  const [searchParams, setSearchParams] = useSearchParams()
  const [creating, setCreating] = useState(false)
  const status = searchParams.get('status') ?? ''
  const severity = searchParams.get('severity') ?? ''
  const active = searchParams.get('active') !== 'false'
  const page = Math.max(1, Number(searchParams.get('page') ?? '1') || 1)

  const list = useListIncidents(
    {
      page,
      pageSize: PAGE_SIZE,
      status: status === '' ? undefined : (status as IncidentStatus),
      severity: severity === '' ? undefined : (severity as Severity),
      active: status === '' && active ? true : undefined,
    },
    { query: { refetchInterval: 15_000 } },
  )
  const data = list.data?.status === 200 ? list.data.data : null
  const totalPages = data === null ? 1 : Math.max(1, Math.ceil(data.total / data.pageSize))

  const update = (name: string, value: string) => {
    const next = new URLSearchParams(searchParams)
    if (value === '') next.delete(name)
    else next.set(name, value)
    next.set('page', '1')
    setSearchParams(next)
  }

  return (
    <section className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Incidentes</h1>
        <button
          type="button"
          onClick={() => setCreating(true)}
          className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800"
        >
          Nuevo incidente
        </button>
      </div>

      {creating && <NewIncidentForm onClose={() => setCreating(false)} />}

      <div className="flex flex-wrap items-center gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
        <select
          aria-label="Estado"
          value={status}
          onChange={(e) => update('status', e.target.value)}
          className="rounded-md border border-slate-300 px-3 py-2 text-sm"
        >
          <option value="">Todos los estados</option>
          {STATUSES.map((s) => (
            <option key={s} value={s}>
              {STATUS_LABELS[s]}
            </option>
          ))}
        </select>
        <select
          aria-label="Severidad"
          value={severity}
          onChange={(e) => update('severity', e.target.value)}
          className="rounded-md border border-slate-300 px-3 py-2 text-sm"
        >
          <option value="">Todas las severidades</option>
          {SEVERITIES.map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>
        {status === '' && (
          <label className="flex items-center gap-2 text-sm">
            <input type="checkbox" checked={active} onChange={(e) => update('active', e.target.checked ? '' : 'false')} />
            Solo activos (abiertos o en curso)
          </label>
        )}
      </div>

      <div className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-50 text-slate-500">
            <tr>
              <th className="px-4 py-3 font-medium">Severidad</th>
              <th className="px-4 py-3 font-medium">Título</th>
              <th className="px-4 py-3 font-medium">Servicio</th>
              <th className="px-4 py-3 font-medium">Estado</th>
              <th className="px-4 py-3 font-medium">Responsable</th>
              <th className="px-4 py-3 font-medium">Abierto</th>
            </tr>
          </thead>
          <tbody>
            {data?.items.map((i) => (
              <tr key={i.id} className="border-t border-slate-100">
                <td className="px-4 py-3">
                  <span className={`rounded px-2 py-0.5 text-xs font-semibold ${SEVERITY_STYLES[i.severity]}`}>
                    {i.severity}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <Link to={`/incidentes/${i.id}`} className="font-medium hover:underline">
                    {i.title}
                  </Link>
                  {i.autoCreated && <span className="ml-2 text-xs text-slate-500">(automático)</span>}
                </td>
                <td className="px-4 py-3">{i.serviceName || '—'}</td>
                <td className="px-4 py-3">
                  <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_STYLES[i.status]}`}>
                    {STATUS_LABELS[i.status]}
                  </span>
                </td>
                <td className="px-4 py-3">{i.assigneeName || 'Sin asignar'}</td>
                <td className="px-4 py-3">{formatDateTime(i.openedAt)}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {data !== null && data.items.length === 0 && (
          <p className="px-4 py-6 text-sm text-slate-500">No hay incidentes con esos filtros.</p>
        )}
      </div>

      {totalPages > 1 && (
        <div className="flex items-center justify-center gap-3 text-sm">
          <button
            type="button"
            disabled={page <= 1}
            onClick={() => update('page', String(page - 1))}
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
            onClick={() => update('page', String(page + 1))}
            className="rounded-md border border-slate-300 px-3 py-1 disabled:opacity-40"
          >
            Siguiente
          </button>
        </div>
      )}
    </section>
  )
}

const NewIncidentForm = ({ onClose }: { onClose: () => void }) => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const create = useCreateIncident()
  const services = useListServices({ pageSize: 100 })
  const serviceItems = services.data?.status === 200 ? services.data.data.items : []
  const [title, setTitle] = useState('')
  const [severity, setSeverity] = useState<Severity>('SEV3')
  const [serviceId, setServiceId] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError(null)
    setErrors({})
    try {
      const result = await create.mutateAsync({
        data: { title, severity, description, serviceId: serviceId === '' ? undefined : serviceId },
      })
      await queryClient.invalidateQueries({ queryKey: getListIncidentsQueryKey() })
      if (result.status === 201) void navigate(`/incidentes/${result.data.id}`)
      else onClose()
    } catch (cause) {
      setErrors(fieldErrors(cause))
      setError(formErrorMessage(cause, 'No se pudo abrir el incidente'))
    }
  }

  return (
    <form
      onSubmit={(e) => void submit(e)}
      className="grid gap-3 rounded-xl border border-slate-200 bg-white p-4 shadow-sm md:grid-cols-4"
    >
      <h2 className="text-lg font-semibold md:col-span-4">Nuevo incidente</h2>
      {error !== null && (
        <p role="alert" className="text-sm text-red-700 md:col-span-4">
          {error}
        </p>
      )}
      <label className="space-y-1 text-sm md:col-span-2">
        <span className="font-medium">Título</span>
        <input
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className="w-full rounded-md border border-slate-300 px-3 py-2"
        />
        {errors.title !== undefined && <span className="text-xs text-red-700">{errors.title}</span>}
      </label>
      <label className="space-y-1 text-sm">
        <span className="font-medium">Severidad</span>
        <select
          value={severity}
          onChange={(e) => setSeverity(e.target.value as Severity)}
          className="w-full rounded-md border border-slate-300 px-3 py-2"
        >
          {SEVERITIES.map((s) => (
            <option key={s} value={s}>
              {s}
            </option>
          ))}
        </select>
      </label>
      <label className="space-y-1 text-sm">
        <span className="font-medium">Servicio afectado</span>
        <select
          value={serviceId}
          onChange={(e) => setServiceId(e.target.value)}
          className="w-full rounded-md border border-slate-300 px-3 py-2"
        >
          <option value="">Ninguno</option>
          {serviceItems.map((s) => (
            <option key={s.id} value={s.id}>
              {s.name}
            </option>
          ))}
        </select>
      </label>
      <label className="space-y-1 text-sm md:col-span-4">
        <span className="font-medium">Descripción</span>
        <textarea
          value={description}
          onChange={(e) => setDescription(e.target.value)}
          rows={3}
          className="w-full rounded-md border border-slate-300 px-3 py-2"
        />
      </label>
      <div className="flex gap-2 md:col-span-4">
        <button type="submit" className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white">
          Abrir incidente
        </button>
        <button type="button" onClick={onClose} className="rounded-md border border-slate-300 px-3 py-2 text-sm">
          Cancelar
        </button>
      </div>
    </form>
  )
}
