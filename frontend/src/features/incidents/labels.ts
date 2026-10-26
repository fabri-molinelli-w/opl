import type { IncidentStatus, Severity } from '@/api/gen/model'

export const SEVERITIES: Severity[] = ['SEV1', 'SEV2', 'SEV3', 'SEV4']

export const SEVERITY_STYLES: Record<Severity, string> = {
  SEV1: 'bg-red-600 text-white',
  SEV2: 'bg-orange-500 text-white',
  SEV3: 'bg-amber-200 text-amber-900',
  SEV4: 'bg-slate-200 text-slate-700',
}

export const STATUSES: IncidentStatus[] = ['open', 'in_progress', 'resolved', 'closed']

export const STATUS_LABELS: Record<IncidentStatus, string> = {
  open: 'Abierto',
  in_progress: 'En curso',
  resolved: 'Resuelto',
  closed: 'Cerrado',
}

export const STATUS_STYLES: Record<IncidentStatus, string> = {
  open: 'bg-red-100 text-red-700',
  in_progress: 'bg-amber-100 text-amber-800',
  resolved: 'bg-green-100 text-green-700',
  closed: 'bg-slate-200 text-slate-600',
}

export const TRANSITION_LABELS: Record<IncidentStatus, string> = {
  open: 'Reabrir',
  in_progress: 'Tomar / En curso',
  resolved: 'Resolver',
  closed: 'Cerrar',
}

export const formatDateTime = (value: string): string =>
  new Intl.DateTimeFormat('es-AR', { dateStyle: 'short', timeStyle: 'short' }).format(new Date(value))
