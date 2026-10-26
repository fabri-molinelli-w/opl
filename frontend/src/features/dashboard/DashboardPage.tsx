import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import {
  Bar,
  BarChart,
  CartesianGrid,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts'
import { useGetDashboardSummary } from '@/api/gen'
import type { DashboardSummary } from '@/api/gen/model'
import { CATEGORY_LABELS, formatDate } from '@/features/documents/categories'
import { SEVERITY_STYLES, STATUS_LABELS } from '@/features/incidents/labels'
import { SystemStatusIndicator } from './SystemStatusIndicator'

const hourMinute = (value: string): string =>
  new Intl.DateTimeFormat('es-AR', { hour: '2-digit', minute: '2-digit' }).format(new Date(value))

type Tone = 'neutral' | 'green' | 'amber' | 'red' | 'slate'

const TONES: Record<Tone, string> = {
  neutral: 'border-slate-200',
  green: 'border-green-300 bg-green-50',
  amber: 'border-amber-300 bg-amber-50',
  red: 'border-red-300 bg-red-50',
  slate: 'border-slate-300 bg-slate-50',
}

const StatCard = ({ label, value, to, tone = 'neutral', hint }: {
  label: string
  value: ReactNode
  to: string
  tone?: Tone
  hint?: string
}) => (
  <Link
    to={to}
    className={`block rounded-xl border bg-white p-4 shadow-sm transition hover:shadow-md ${TONES[tone]}`}
  >
    <p className="text-sm text-slate-500">{label}</p>
    <p className="mt-1 text-3xl font-semibold">{value}</p>
    {hint !== undefined && <p className="mt-1 text-xs text-slate-500">{hint}</p>}
  </Link>
)

const Panel = ({ title, children }: { title: string; children: ReactNode }) => (
  <div className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
    <h2 className="mb-3 text-sm font-semibold uppercase tracking-wide text-slate-500">{title}</h2>
    {children}
  </div>
)

export const DashboardPage = () => {
  // RN-1: un solo request; refresco automático cada 30 s sin recargar (CA-2).
  const summary = useGetDashboardSummary({ query: { refetchInterval: 30_000 } })
  const data = summary.data?.status === 200 ? summary.data.data : null

  return (
    <section className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-2xl font-semibold">Dashboard</h1>
        <SystemStatusIndicator />
      </div>
      {summary.isLoading && <p className="text-sm text-slate-500">Cargando…</p>}
      {summary.isError && (
        <p role="alert" className="text-sm text-red-700">
          No se pudo cargar el resumen.
        </p>
      )}
      {data !== null && <DashboardContent data={data} />}
    </section>
  )
}

const DashboardContent = ({ data }: { data: DashboardSummary }) => {
  const { servers, services, documents, incidents, availability } = data
  const categoryData = documents.byCategory.map((c) => ({ name: CATEGORY_LABELS[c.category], total: c.count }))
  const series = availability.series.map((b) => ({ ...b, label: hourMinute(b.start) }))

  return (
    <>
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard
          label="Servicios caídos"
          value={services.down}
          to="/inventario"
          tone={services.down > 0 ? 'red' : 'green'}
          hint={`${services.up} UP · ${services.degraded} DEGRADED · ${services.unknown} sin datos`}
        />
        <StatCard
          label="Incidentes activos"
          value={incidents.active}
          to="/incidentes"
          tone={incidents.sev1 + incidents.sev2 > 0 ? 'red' : incidents.active > 0 ? 'amber' : 'green'}
          hint={`SEV1 ${incidents.sev1} · SEV2 ${incidents.sev2} · SEV3 ${incidents.sev3} · SEV4 ${incidents.sev4}`}
        />
        <StatCard
          label="Disponibilidad 24 h"
          value={availability.uptimePercent24h !== undefined ? `${availability.uptimePercent24h}%` : '—'}
          to="/inventario"
          tone="slate"
          hint={`${availability.checks24h} chequeos`}
        />
        <StatCard
          label="Documentos"
          value={documents.total}
          to="/documentacion"
          hint={`${documents.updatedLast7Days} actualizados en los últimos 7 días`}
        />
      </div>

      <div className="grid gap-4 sm:grid-cols-3 xl:grid-cols-6">
        <StatCard label="Servidores" value={servers.total} to="/inventario" />
        <StatCard label="En prod" value={servers.prod} to="/inventario?environment=prod" />
        <StatCard label="En staging" value={servers.staging} to="/inventario?environment=staging" />
        <StatCard label="En dev" value={servers.dev} to="/inventario?environment=dev" />
        <StatCard label="On-premise" value={servers.onprem} to="/inventario" />
        <StatCard label="Cloud" value={servers.cloud} to="/inventario" />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Panel title="Servicios con problemas">
          {services.problems.length === 0 ? (
            <p className="text-sm text-slate-500">Todos los servicios monitoreados responden.</p>
          ) : (
            <ul className="space-y-2 text-sm">
              {services.problems.map((p) => (
                <li key={p.id} className="flex items-center justify-between gap-2">
                  <Link to={`/inventario/${p.serverId}`} className="hover:underline">
                    {p.name} <span className="text-slate-500">· {p.hostname}</span>
                  </Link>
                  <span
                    className={`rounded-full px-2 py-0.5 text-xs font-medium ${
                      p.status === 'DOWN' ? 'bg-red-100 text-red-700' : 'bg-amber-100 text-amber-700'
                    }`}
                  >
                    {p.status}
                  </span>
                </li>
              ))}
            </ul>
          )}
        </Panel>
        <Panel title="Incidentes activos">
          {incidents.recent.length === 0 ? (
            <p className="text-sm text-slate-500">No hay incidentes abiertos.</p>
          ) : (
            <ul className="space-y-2 text-sm">
              {incidents.recent.map((i) => (
                <li key={i.id} className="flex items-center gap-2">
                  <span className={`rounded px-1.5 py-0.5 text-xs font-semibold ${SEVERITY_STYLES[i.severity]}`}>
                    {i.severity}
                  </span>
                  <Link to={`/incidentes/${i.id}`} className="flex-1 hover:underline">
                    {i.title}
                  </Link>
                  <span className="text-xs text-slate-500">{STATUS_LABELS[i.status]}</span>
                </li>
              ))}
            </ul>
          )}
        </Panel>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Panel title="Disponibilidad y latencia (24 h)">
          {series.length === 0 ? (
            <p className="text-sm text-slate-500">Todavía no hay chequeos en las últimas 24 h.</p>
          ) : (
            <div className="h-56">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={series}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis dataKey="label" fontSize={12} />
                  <YAxis yAxisId="pct" domain={[0, 100]} fontSize={12} unit="%" />
                  <YAxis yAxisId="ms" orientation="right" fontSize={12} unit=" ms" />
                  <Tooltip />
                  <Line yAxisId="pct" dataKey="uptimePercent" name="Disponibilidad" stroke="#16a34a" dot={false} />
                  <Line yAxisId="ms" dataKey="avgResponseMs" name="Latencia" stroke="#0369a1" dot={false} />
                </LineChart>
              </ResponsiveContainer>
            </div>
          )}
        </Panel>
        <Panel title="Documentos por categoría">
          {categoryData.length === 0 ? (
            <p className="text-sm text-slate-500">Todavía no hay documentos.</p>
          ) : (
            <div className="h-56">
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={categoryData}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#e2e8f0" />
                  <XAxis dataKey="name" fontSize={12} />
                  <YAxis allowDecimals={false} fontSize={12} />
                  <Tooltip />
                  <Bar dataKey="total" name="Documentos" fill="#334155" radius={[4, 4, 0, 0]} />
                </BarChart>
              </ResponsiveContainer>
            </div>
          )}
        </Panel>
      </div>

      <Panel title="Últimos documentos editados">
        {documents.recent.length === 0 ? (
          <p className="text-sm text-slate-500">Todavía no hay documentos.</p>
        ) : (
          <ul className="divide-y divide-slate-100 text-sm">
            {documents.recent.map((d) => (
              <li key={d.id} className="flex items-center justify-between py-2">
                <Link to={`/documentacion/${d.id}`} className="font-medium hover:underline">
                  {d.title}
                </Link>
                <span className="text-xs text-slate-500">
                  {d.updatedByName} · {formatDate(d.updatedAt)}
                </span>
              </li>
            ))}
          </ul>
        )}
      </Panel>
    </>
  )
}
