import { useState, type FormEvent, type ReactNode } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import {
  getGetServerQueryKey,
  getListServersQueryKey,
  useCreateServer,
  useUpdateServer,
} from '@/api/gen'
import type { Environment, Location, Server } from '@/api/gen/model'
import { fieldErrors, formErrorMessage } from './formErrors'

const ENVIRONMENTS: Environment[] = ['prod', 'staging', 'dev']
const LOCATIONS: Location[] = ['onprem', 'cloud']

type Props = {
  server?: Server
  onClose: () => void
}

export const ServerForm = ({ server, onClose }: Props) => {
  const queryClient = useQueryClient()
  const createServer = useCreateServer()
  const updateServer = useUpdateServer()
  const [hostname, setHostname] = useState(server?.hostname ?? '')
  const [ip, setIp] = useState(server?.ip ?? '')
  const [environment, setEnvironment] = useState<Environment>(
    server?.environment ?? 'prod',
  )
  const [os, setOs] = useState(server?.os ?? '')
  const [location, setLocation] = useState<Location>(
    server?.location ?? 'onprem',
  )
  const [owner, setOwner] = useState(server?.owner ?? '')
  const [tags, setTags] = useState(server?.tags.join(', ') ?? '')
  const [error, setError] = useState<string | null>(null)
  const [errors, setErrors] = useState<Record<string, string>>({})

  const submit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setError(null)
    setErrors({})
    const data = {
      hostname,
      ip,
      environment,
      os,
      location,
      owner,
      tags: tags
        .split(',')
        .map((tag) => tag.trim())
        .filter((tag) => tag.length > 0),
    }
    try {
      if (server === undefined) {
        await createServer.mutateAsync({ data })
      } else {
        await updateServer.mutateAsync({ id: server.id, data })
      }
      await queryClient.invalidateQueries({ queryKey: getListServersQueryKey() })
      if (server !== undefined) {
        await queryClient.invalidateQueries({
          queryKey: getGetServerQueryKey(server.id),
        })
      }
      onClose()
    } catch (cause) {
      setErrors(fieldErrors(cause))
      setError(formErrorMessage(cause, 'No se pudo guardar el servidor'))
    }
  }

  return (
    <form
      onSubmit={(event) => void submit(event)}
      className="space-y-4 rounded-xl border border-slate-200 bg-white p-4 shadow-sm"
    >
      <h2 className="text-lg font-semibold">
        {server === undefined ? 'Nuevo servidor' : `Editar ${server.hostname}`}
      </h2>
      {error !== null && (
        <p role="alert" className="text-sm text-red-600">
          {error}
        </p>
      )}
      <div className="grid gap-4 md:grid-cols-2">
        <Field label="Hostname" error={errors.hostname}>
          <input
            required
            value={hostname}
            onChange={(event) => setHostname(event.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          />
        </Field>
        <Field label="IP" error={errors.ip}>
          <input
            required
            value={ip}
            onChange={(event) => setIp(event.target.value)}
            placeholder="192.0.2.10"
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          />
        </Field>
        <Field label="Environment" error={errors.environment}>
          <select
            value={environment}
            onChange={(event) =>
              setEnvironment(event.target.value as Environment)
            }
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          >
            {ENVIRONMENTS.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Location" error={errors.location}>
          <select
            value={location}
            onChange={(event) => setLocation(event.target.value as Location)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          >
            {LOCATIONS.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </Field>
        <Field label="Sistema operativo" error={errors.os}>
          <input
            value={os}
            onChange={(event) => setOs(event.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          />
        </Field>
        <Field label="Responsable" error={errors.owner}>
          <input
            value={owner}
            onChange={(event) => setOwner(event.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          />
        </Field>
        <Field label="Tags (separados por coma)" error={errors.tags}>
          <input
            value={tags}
            onChange={(event) => setTags(event.target.value)}
            className="w-full rounded-md border border-slate-300 px-3 py-2"
          />
        </Field>
      </div>
      <div className="flex gap-3">
        <button
          type="submit"
          disabled={createServer.isPending || updateServer.isPending}
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
