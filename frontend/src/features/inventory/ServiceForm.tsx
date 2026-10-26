import { useState, type FormEvent, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import {
  getGetServerQueryKey,
  useCreateService,
  useUpdateService,
} from '@/api/gen'
import type { CheckType, Service, ServiceType } from '@/api/gen/model'
import { fieldErrors, formErrorMessage } from './formErrors'

const SERVICE_TYPES: ServiceType[] = ['web', 'db', 'queue', 'other']
const CHECK_TYPES: CheckType[] = ['http', 'tcp']

type Props = {
  serverId: string
  service?: Service
  onClose: () => void
}

export const ServiceForm = ({ serverId, service, onClose }: Props) => {
  const queryClient = useQueryClient()
  const createService = useCreateService()
  const updateService = useUpdateService()
  const [name, setName] = useState(service?.name ?? '')
  const [type, setType] = useState<ServiceType>(service?.type ?? 'web')
  const [port, setPort] = useState(String(service?.port ?? 80))
  const [checkType, setCheckType] = useState<CheckType>(
    service?.checkType ?? 'http',
  )
  const [checkTarget, setCheckTarget] = useState(service?.checkTarget ?? '')
  const [expectedStatus, setExpectedStatus] = useState(
    String(service?.expectedStatus ?? 200),
  )
  const [intervalSeconds, setIntervalSeconds] = useState(
    String(service?.intervalSeconds ?? 30),
  )
  const [latencyThresholdMs, setLatencyThresholdMs] = useState(
    String(service?.latencyThresholdMs ?? 1000),
  )
  const [error, setError] = useState<string | null>(null)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError(null)
    setErrors({})
    const data = {
      name,
      type,
      port: Number(port),
      checkType,
      checkTarget,
      expectedStatus: Number(expectedStatus),
      intervalSeconds: Number(intervalSeconds),
      latencyThresholdMs: Number(latencyThresholdMs),
    }
    try {
      if (service === undefined) {
        await createService.mutateAsync({ data: { serverId, ...data } })
      } else {
        await updateService.mutateAsync({ id: service.id, data })
      }
      await queryClient.invalidateQueries({
        queryKey: getGetServerQueryKey(serverId),
      })
      onClose()
    } catch (cause) {
      setErrors(fieldErrors(cause))
      setError(formErrorMessage(cause, 'No se pudo guardar el servicio'))
    }
  }

  return (
    <form
      onSubmit={(event) => void submit(event)}
      className="space-y-4 rounded-xl border border-slate-200 bg-white p-4 shadow-sm"
    >
      <h2 className="text-lg font-semibold">
        {service === undefined ? 'Nuevo servicio' : `Editar ${service.name}`}
      </h2>
      {error !== null && (
        <p role="alert" className="text-sm text-red-600">
          {error}
        </p>
      )}
      <div className="grid gap-4 md:grid-cols-2">
        <Field label="Nombre" error={errors.name}>
          <input
            required
            value={name}
            onChange={(event) => setName(event.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          />
        </Field>
        <Field label="Tipo" error={errors.type}>
          <select
            value={type}
            onChange={(event) => setType(event.target.value as ServiceType)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          >
            {SERVICE_TYPES.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Puerto" error={errors.port}>
          <input
            required
            type="number"
            min={1}
            max={65535}
            value={port}
            onChange={(event) => setPort(event.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          />
        </Field>
        <Field label="Tipo de check" error={errors.checkType}>
          <select
            value={checkType}
            onChange={(event) => setCheckType(event.target.value as CheckType)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          >
            {CHECK_TYPES.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Check target" error={errors.checkTarget}>
          <input
            required
            value={checkTarget}
            onChange={(event) => setCheckTarget(event.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          />
        </Field>
        <Field label="Status esperado" error={errors.expectedStatus}>
          <input
            type="number"
            min={100}
            max={599}
            value={expectedStatus}
            onChange={(event) => setExpectedStatus(event.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          />
        </Field>
        <Field label="Intervalo (segundos)" error={errors.intervalSeconds}>
          <input
            type="number"
            min={1}
            value={intervalSeconds}
            onChange={(event) => setIntervalSeconds(event.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          />
        </Field>
        <Field label="Umbral de latencia (ms)" error={errors.latencyThresholdMs}>
          <input
            type="number"
            min={1}
            value={latencyThresholdMs}
            onChange={(event) => setLatencyThresholdMs(event.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          />
        </Field>
      </div>
      <div className="flex gap-3">
        <button
          type="submit"
          disabled={createService.isPending || updateService.isPending}
          className="rounded-md bg-slate-900 px-3 py-2 text-sm font-medium text-white hover:bg-slate-800 disabled:opacity-50"
        >
          Guardar
        </button>
        <button
          type="button"
          onClick={onClose}
          className="rounded-md border border-slate-300 px-3 py-2 text-sm font-medium"
        >
          Cancelar
        </button>
      </div>
    </form>
  )
}

const Field = ({
  label,
  error,
  children,
}: {
  label: string
  error?: string
  children: ReactNode
}) => (
  <label className="block space-y-1 text-sm">
    <span className="font-medium text-slate-700">{label}</span>
    {children}
    {error !== undefined && (
      <span className="block text-xs text-red-600">{error}</span>
    )}
  </label>
)
