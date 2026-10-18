import { useGetHealth } from '@/api/gen/system/system'
import type { Health } from '@/api/gen/model'

type SystemState = 'up' | 'degraded' | 'offline'

const STATUS_STYLES: Record<SystemState, { dot: string; label: string }> = {
  up: { dot: 'bg-green-500', label: 'Operativo' },
  degraded: { dot: 'bg-red-500', label: 'Degradado' },
  offline: { dot: 'bg-gray-400', label: 'Sin conexión' },
}

const deriveState = (
  health: Health | undefined,
  isError: boolean,
): SystemState => {
  if (isError || health === undefined) {
    return 'offline'
  }
  if (health.status === 'ok' && health.database === 'up') {
    return 'up'
  }
  return 'degraded'
}

export const SystemStatusIndicator = () => {
  const { data, isError, isPending } = useGetHealth({
    query: { refetchInterval: 30_000 },
  })

  const health: Health | undefined = data?.data
  const state: SystemState = isPending ? 'offline' : deriveState(health, isError)
  const { dot, label } = STATUS_STYLES[state]

  return (
    <div className="inline-flex items-center gap-3 rounded-lg border border-slate-200 bg-white px-5 py-4 shadow-sm">
      <span className={`h-3 w-3 rounded-full ${dot}`} aria-hidden="true" />
      <div>
        <p className="text-sm text-slate-500">Estado del sistema</p>
        <p className="text-base font-medium">{label}</p>
      </div>
    </div>
  )
}
