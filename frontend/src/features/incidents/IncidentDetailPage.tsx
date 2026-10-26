import { useState, type FormEvent } from 'react'
import { Link, useParams } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import {
  getGetIncidentQueryKey,
  getListIncidentsQueryKey,
  useAddIncidentNote,
  useGetIncident,
  useTransitionIncident,
  useUpdateIncident,
} from '@/api/gen'
import type { Incident, IncidentStatus } from '@/api/gen/model'
import { isApiError } from '@/api/http'
import { useAuth } from '@/auth/useAuth'
import { SEVERITY_STYLES, STATUS_LABELS, STATUS_STYLES, TRANSITION_LABELS, formatDateTime } from './labels'

const EVENT_STYLES: Record<string, string> = {
  status_change: 'border-sky-400',
  note: 'border-slate-300',
  system: 'border-amber-400',
}

export const IncidentDetailPage = () => {
  const { id = '' } = useParams()
  const result = useGetIncident(id, { query: { refetchInterval: 15_000 } })
  const incident = result.data?.status === 200 ? result.data.data : null

  if (result.isLoading) return <p className="text-sm text-slate-500">Cargando incidente…</p>
  if (incident === null) {
    return (
      <p role="alert" className="text-sm text-red-700">
        El incidente no existe.
      </p>
    )
  }
  return <IncidentDetail incident={incident} />
}

const IncidentDetail = ({ incident }: { incident: Incident }) => {
  const { user } = useAuth()
  const queryClient = useQueryClient()
  const transition = useTransitionIncident()
  const addNote = useAddIncidentNote()
  const updateIncident = useUpdateIncident()
  const [rootCause, setRootCause] = useState(incident.rootCause)
  const [note, setNote] = useState('')
  const [error, setError] = useState<string | null>(null)

  const refresh = async () => {
    await queryClient.invalidateQueries({ queryKey: getGetIncidentQueryKey(incident.id) })
    await queryClient.invalidateQueries({ queryKey: getListIncidentsQueryKey() })
  }

  const run = async (action: () => Promise<unknown>) => {
    setError(null)
    try {
      await action()
      await refresh()
    } catch (cause) {
      setError(isApiError(cause) ? cause.message : 'No se pudo completar la acción')
    }
  }

  const move = (to: IncidentStatus) =>
    run(() =>
      transition.mutateAsync({
        id: incident.id,
        data: { to, rootCause: to === 'resolved' ? rootCause : undefined },
      }),
    )

  const submitNote = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    void run(async () => {
      await addNote.mutateAsync({ id: incident.id, data: { body: note } })
      setNote('')
    })
  }

  const canResolve = incident.allowedTransitions.includes('resolved')

  return (
    <section className="grid gap-6 lg:grid-cols-[1fr_22rem]">
      <div className="space-y-4">
        <Link to="/incidentes" className="text-sm text-sky-700 hover:underline">
          ← Incidentes
        </Link>
        <div className="space-y-3 rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <div className="flex flex-wrap items-center gap-2">
            <span className={`rounded px-2 py-0.5 text-xs font-semibold ${SEVERITY_STYLES[incident.severity]}`}>
              {incident.severity}
            </span>
            <span className={`rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_STYLES[incident.status]}`}>
              {STATUS_LABELS[incident.status]}
            </span>
            {incident.autoCreated && <span className="text-xs text-slate-500">Abierto automáticamente</span>}
          </div>
          <h1 className="text-2xl font-semibold">{incident.title}</h1>
          {incident.description !== '' && <p className="whitespace-pre-wrap text-slate-700">{incident.description}</p>}
          <dl className="grid gap-3 text-sm sm:grid-cols-2">
            <Item label="Servicio" value={incident.serviceName || '—'} />
            <Item label="Responsable" value={incident.assigneeName || 'Sin asignar'} />
            <Item label="Abierto" value={formatDateTime(incident.openedAt)} />
            <Item
              label="Tiempo de resolución"
              value={incident.resolutionMinutes !== undefined ? `${incident.resolutionMinutes} min` : '—'}
            />
            {incident.rootCause !== '' && <Item label="Causa raíz" value={incident.rootCause} />}
          </dl>
        </div>

        <div className="space-y-3 rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
          <h2 className="font-semibold">Línea de tiempo</h2>
          <ol className="space-y-3">
            {incident.events.map((e) => (
              <li key={e.id} className={`border-l-4 pl-3 ${EVENT_STYLES[e.kind] ?? 'border-slate-300'}`}>
                <p className="text-sm">{e.body}</p>
                <p className="text-xs text-slate-500">
                  {e.authorName} · {formatDateTime(e.createdAt)}
                </p>
              </li>
            ))}
          </ol>
          {incident.status !== 'closed' && (
            <form onSubmit={submitNote} className="flex gap-2">
              <input
                aria-label="Nota"
                value={note}
                onChange={(e) => setNote(e.target.value)}
                placeholder="Agregar una nota…"
                className="flex-1 rounded-md border border-slate-300 px-3 py-2 text-sm"
              />
              <button
                type="submit"
                disabled={note.trim() === ''}
                className="rounded-md bg-slate-900 px-3 py-2 text-sm text-white disabled:opacity-40"
              >
                Agregar
              </button>
            </form>
          )}
        </div>
      </div>

      <aside className="space-y-3 self-start rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
        <h2 className="text-sm font-semibold uppercase tracking-wide text-slate-500">Acciones</h2>
        {error !== null && (
          <p role="alert" className="rounded-md bg-red-50 px-3 py-2 text-sm text-red-800">
            {error}
          </p>
        )}
        {user !== null && incident.assigneeId !== user.id && incident.status !== 'closed' && (
          <button
            type="button"
            onClick={() => void run(() => updateIncident.mutateAsync({ id: incident.id, data: { assigneeId: user.id } }))}
            className="w-full rounded-md border border-slate-300 px-3 py-2 text-sm hover:bg-slate-50"
          >
            Asignármelo
          </button>
        )}
        {canResolve && (
          <label className="block space-y-1 text-sm">
            <span className="font-medium">Causa raíz (obligatoria para resolver)</span>
            <textarea
              value={rootCause}
              onChange={(e) => setRootCause(e.target.value)}
              rows={3}
              className="w-full rounded-md border border-slate-300 px-3 py-2"
            />
          </label>
        )}
        {incident.allowedTransitions.map((to) => (
          <button
            key={to}
            type="button"
            onClick={() => void move(to)}
            className="w-full rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800"
          >
            {TRANSITION_LABELS[to]}
          </button>
        ))}
        {incident.allowedTransitions.length === 0 && (
          <p className="text-sm text-slate-500">El incidente está cerrado.</p>
        )}
      </aside>
    </section>
  )
}

const Item = ({ label, value }: { label: string; value: string }) => (
  <div>
    <dt className="text-xs uppercase tracking-wide text-slate-500">{label}</dt>
    <dd className="mt-0.5 text-slate-800">{value}</dd>
  </div>
)
